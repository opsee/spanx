package service

import (
	"encoding/json"
	"errors"
	"net"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opsee/basic/schema/aws/credentials"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/spanx/roler"
	"github.com/opsee/spanx/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	grpcauth "google.golang.org/grpc/credentials"
)

var (
	errUnauthorized       = errors.New("unauthorized.")
	errMissingAccessKey   = errors.New("missing AccessKeyID.")
	errMissingSecretKey   = errors.New("missing SecretAccessKey.")
	errMissingRegion      = errors.New("missing region.")
	errUnknown            = errors.New("unknown error.")
	errSavingRole         = errors.New("Error saving the Opsee role in your AWS account, please check that you have the necessary permissions.")
	errGettingCredentials = errors.New("Error fetching credentials from AWS STS.")
)

type service struct {
	db store.Store
}

func New(db store.Store) *service {
	return &service{db}
}

func (s *service) Start(listenAddr, cert, certkey string) error {
	auth, err := grpcauth.NewServerTLSFromFile(cert, certkey)
	if err != nil {
		return err
	}

	server := grpc.NewServer(grpc.Creds(auth))
	opsee.RegisterSpanxServer(server, s)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	s.startSQSListener()
	return server.Serve(lis)
}

/*
{
  "Type" : "Notification",
  "MessageId" : "4d12cf8d-821d-5b95-8452-801d2f79ab47",
  "TopicArn" : "arn:aws:sns:us-west-2:933693344490:opsee-cfn-callback",
  "Subject" : "AWS CloudFormation custom resource request",
  "Message" : "{\"RequestType\":\"Create\",\"ServiceToken\":\"arn:aws:sns:us-west-2:933693344490:opsee-cfn-callback\",\"ResponseURL\":\"https://cloudformation-custom-resource-response-uswest2.s3-us-west-2.amazonaws.com/arn%3Aaws%3Acloudformation%3Aus-west-2%3A933693344490%3Astack/opsee-role-stack-greg-test-blah/55c35420-067a-11e6-91a2-503f20f2ad1e%7COpseeNotification%7Ccb46085e-b829-4567-81b8-9f5dd0755c2a?AWSAccessKeyId=AKIAI4KYMPPRGIACET5Q&Expires=1461110596&Signature=WW7JkaF9Nz3jxH90anYrSYZmTww%3D\",\"StackId\":\"arn:aws:cloudformation:us-west-2:933693344490:stack/opsee-role-stack-greg-test-blah/55c35420-067a-11e6-91a2-503f20f2ad1e\",\"RequestId\":\"cb46085e-b829-4567-81b8-9f5dd0755c2a\",\"LogicalResourceId\":\"OpseeNotification\",\"ResourceType\":\"Custom::OpseeNotificationResource\",\"ResourceProperties\":{\"ServiceToken\":\"arn:aws:sns:us-west-2:933693344490:opsee-cfn-callback\",\"RoleExternalID\":\"{{ .User.ExternalID }}\",\"RoleARN\":\"arn:aws:iam::933693344490:role/opsee-role-stack-greg-test-blah-OpseeRole-1ULQTXA2L7RLE\",\"StackName\":\"opsee-role-stack-greg-test-blah\",\"StackID\":\"arn:aws:cloudformation:us-west-2:933693344490:stack/opsee-role-stack-greg-test-blah/55c35420-067a-11e6-91a2-503f20f2ad1e\"}}",
  "Timestamp" : "2016-04-19T22:03:16.248Z",
  "SignatureVersion" : "1",
  "Signature" : "T1F35l+J4GuUt9VDqk5JAd9c14M0sPJrBSFrbyPeXPiXBdLZov5X7/eywPqISvD6oQXR8p0I5MVMxPPNfINNS+nV8sfJ3AEqtMV+XGN3AFDg4Z8yPHmbd4w90i7XNQSLXjsj7BrE6TwLHysJFQ7bspo9agN0pvyoeSjwv+8Jawhea+wq46jRgnl/UcUIn1G3a2P0qYmzrkVvmVN5mBXWjllCGUY0VxHtmU9/ZQOK9jk3n4NyterYq3p5FQDN61URzog07jx5XaOchXblaF4EwT9mKyn8Yg/dQ6wFHsIpB7GsPg89UJPkdmmPp7fAezALGfD+sBYixgUYN1w77wgK8A==",
  "SigningCertURL" : "https://sns.us-west-2.amazonaws.com/SimpleNotificationService-bb750dd426d95ee9390147a5624348ee.pem",
  "UnsubscribeURL" : "https://sns.us-west-2.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-west-2:933693344490:opsee-cfn-callback:0263cdc7-1a80-4c62-835d-e422fc931962"
}
*/

func (s *service) startSQSListener() {
	var region string

	region = os.Getenv("AWS_DEFAULT_REGION")

	if region == "" {
		region = os.Getenv("AWS_REGION")
	}

	if region == "" {
		mdc := ec2metadata.New(session.New())
		if mdc.Available() {
			region, _ = mdc.Region()
		}
	}

	sqsClient := sqs.New(session.New(aws.NewConfig().WithRegion(region)))
	NewPoller(sqsClient,
		"https://sqs.us-west-2.amazonaws.com/933693344490/opsee-cfn-callback",
		func(msg string) error {
			notif := &CFNotification{}

			if err := json.Unmarshal([]byte(msg), notif); err != nil {
				return err
			}
			return s.handleCFNCallback(notif)
		}).Poll()
}

// EnhancedCombatMode returns a URL to a CFN template in a specified region
// that a customer can use to launch a role.
func (s *service) EnhancedCombatMode(ctx context.Context, req *opsee.EnhancedCombatModeRequest) (*opsee.EnhancedCombatModeResponse, error) {
	if req.Region == "" {
		return nil, errMissingRegion
	}

	if req.User == nil {
		return nil, errUnauthorized
	}

	log.WithFields(log.Fields{
		"customer_id": req.User.CustomerId,
		"endpoint":    "EnhancedCombatMode",
	}).Info("grpc request")

	url, err := roler.GetStackURL(s.db, req.User.CustomerId, req.Region)
	if err != nil {
		log.WithFields(log.Fields{
			"customer_id": req.User.CustomerId,
			"endpoint":    "EnhancedCombatMode",
		}).WithError(err).Error("Error getting URL for customer.")
		return nil, errors.New("Error getting URL for customer.")
	}

	return &opsee.EnhancedCombatModeResponse{url}, nil
}

func (s *service) PutRole(ctx context.Context, req *opsee.PutRoleRequest) (*opsee.PutRoleResponse, error) {
	log.WithFields(log.Fields{
		"customer_id": req.User.CustomerId,
		"endpoint":    "PutRole",
	}).Info("grpc request")

	creds, err := roler.ResolveCredentials(s.db, req.User.CustomerId, req.Credentials.GetAccessKeyID(), req.Credentials.GetSecretAccessKey())
	if err != nil {
		return nil, errSavingRole
	}

	return &opsee.PutRoleResponse{
		Credentials: &credentials.Value{
			AccessKeyID:     aws.String(creds.AccessKeyID),
			SecretAccessKey: aws.String(creds.SecretAccessKey),
			SessionToken:    aws.String(creds.SessionToken),
		},
	}, nil
}

func (s *service) GetCredentials(ctx context.Context, req *opsee.GetCredentialsRequest) (*opsee.GetCredentialsResponse, error) {
	log.WithFields(log.Fields{
		"customer_id": req.User.CustomerId,
		"endpoint":    "GetCredentials",
	}).Info("grpc request")

	creds, err := roler.GetCredentials(s.db, req.User.CustomerId)
	if err != nil {
		return nil, errGettingCredentials
	}

	return &opsee.GetCredentialsResponse{
		Credentials: &credentials.Value{
			AccessKeyID:     aws.String(creds.AccessKeyID),
			SecretAccessKey: aws.String(creds.SecretAccessKey),
			SessionToken:    aws.String(creds.SessionToken),
		},
	}, nil
}
