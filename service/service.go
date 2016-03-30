package service

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/opsee/basic/schema/aws/credentials"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/spanx/roler"
	"github.com/opsee/spanx/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	grpcauth "google.golang.org/grpc/credentials"
	"net"
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

	return server.Serve(lis)
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
