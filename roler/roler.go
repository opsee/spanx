package roler

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cenkalti/backoff"
	"github.com/opsee/basic/com"
	"github.com/opsee/spanx/policies"
	"github.com/opsee/spanx/store"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	awsSession *session.Session
	arnRegexp  = regexp.MustCompile(`^arn:aws:iam::(\d+):(user.+|root)$`)

	AccountNotFound         = errors.New("AWS account for that customer not found.")
	InsufficientPermissions = fmt.Errorf("IAM role or user provided has insufficient permissions to provision a role. The minimum policy required to launch Opsee is:\n%s", policies.UserPolicy)
)

func init() {
	var (
		ec2meta *ec2metadata.EC2Metadata
		region  string
		creds   *credentials.Credentials
	)

	ec2meta = ec2metadata.New(session.New())
	if ec2meta.Available() {
		// ignoring error here since we'll try to get region from env later
		region, _ = ec2meta.Region()
	}

	if region == "" {
		region = getEnvRegion()
	}

	creds = credentials.NewChainCredentials(
		[]credentials.Provider{
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2meta,
			},
			&credentials.EnvProvider{},
		},
	)

	awsSession = session.New(&aws.Config{
		Credentials: creds,
		MaxRetries:  aws.Int(11),
		Region:      aws.String(region),
	})
}

func ResolveCredentials(db store.Store, customerID, accessKey, secretKey string) (credentials.Value, error) {
	var (
		account *com.Account
		creds   credentials.Value
		err     error
		logger  = log.WithFields(log.Fields{"customer_id": customerID})
	)

	// grab the account id and persist it in an account object
	iamClient := iam.New(session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
		MaxRetries: aws.Int(11),
		Region:     aws.String(getEnvRegion()),
	}))

	user, err := iamClient.GetUser(nil)
	if err != nil {
		return creds, handleAWSError("GetUser", err)
	}

	if user.User == nil {
		err = fmt.Errorf("No user returned from aws sdk, this shouldn't happen")
		logger.WithError(err).Error("error fetching user")
		return creds, err
	}

	arn := aws.StringValue(user.User.Arn)
	if arn == "" {
		err = fmt.Errorf("No user found when fetching the current user from provided credentials")
		logger.WithError(err).Error("error fetching user ARN")
		return creds, err
	}

	accountID, err := parseARNAccount(arn)
	if err != nil {
		logger.WithError(err).Error("error parsing user ARN")
		return creds, err
	}

	account = &com.Account{
		ID:         accountID,
		CustomerID: customerID,
		Active:     true,
	}

	err = resolveAccount(db, account)
	if err != nil {
		logger.WithError(err).Error("error resolving account")
		return creds, err
	}

	// time 2 provision a policy / role for us in their aws account
	_, err = iamClient.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(fmt.Sprintf(policies.AssumeRolePolicy, customerID)),
		RoleName:                 aws.String(account.RoleName()),
	})

	if err = handleAWSError("CreateRole", err); err != nil {
		logger.WithError(err).Error("error creating role")
		return creds, err
	}

	_, err = iamClient.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policies.GetPolicy()),
		PolicyName:     aws.String(account.PolicyName()),
		RoleName:       aws.String(account.RoleName()),
	})

	if err = handleAWSError("PutRolePolicy", err); err != nil {
		logger.WithError(err).Error("error putting role policy")
		return creds, err
	}

	// go ahead and return some creds
	return getAccountCredentials(db, account)
}

func AddExternalId(db store.Store, customerID string, externalID string) (*com.Account, error) {
	account, err := db.GetAccount(&store.GetAccountRequest{CustomerID: customerID, Active: true})
	if err != nil {
		log.WithFields(log.Fields{"customer_id": customerID}).WithError(err).Error("error getting account from db")
		return nil, err
	}

	account.ExternalID = externalID

	err = store.UpdateAccount(account, account)
	if err != nil {
		log.WithFields(log.Fields{"customer_id": customerID}).WithError(err).Error("error getting account from db")
		return nil, err
	}

	return account, nil

}

func GetCredentials(db store.Store, customerID string) (credentials.Value, error) {
	account, err := db.GetAccount(&store.GetAccountRequest{CustomerID: customerID, Active: true})
	if err != nil {
		log.WithFields(log.Fields{"customer_id": customerID}).WithError(err).Error("error getting account from db")
		return credentials.Value{}, err
	}

	return getAccountCredentials(db, account)
}

func DeleteCredentials(db store.Store, customerID string) error {
	account, err := db.GetAccount(&store.GetAccountRequest{CustomerID: customerID, Active: true})
	if err != nil {
		log.WithFields(log.Fields{"customer_id": customerID}).WithError(err).Error("error getting account from db")
		return err
	}

	return db.DeleteAccount(account)
}

func resolveAccount(db store.Store, account *com.Account) error {
	// find out if we already have an account saved
	oldAccount, err := db.GetAccount(&store.GetAccountRequest{CustomerID: account.CustomerID, Active: true})
	if err != nil {
		return err
	}

	// no previous account, just make a new one
	if oldAccount == nil {
		return db.PutAccount(account)
	}

	// a previous account, but not the same
	if oldAccount.ID != account.ID {
		return db.UpdateAccount(oldAccount, account)
	}

	return nil
}

func getAccountCredentials(db store.Store, account *com.Account) (credentials.Value, error) {
	var (
		creds credentials.Value
		err   error
	)

	if account != nil {
		backoff.Retry(func() error {
			var arn string

			if account.RoleARN != "" {
				arn = account.RoleARN
			} else {
				arn = account.ComputedRoleARN()
			}

			creds, err = stscreds.NewCredentials(awsSession, arn, func(arp *stscreds.AssumeRoleProvider) {
				arp.ExternalID = aws.String(account.CustomerID)
				arp.Duration = 60 * time.Minute
			}).Get()

			if err != nil {
				return err
			}

			err = nil
			return nil

		}, &backoff.ExponentialBackOff{
			InitialInterval:     100 * time.Millisecond,
			RandomizationFactor: 0.5,
			Multiplier:          1.5,
			MaxInterval:         time.Second,
			MaxElapsedTime:      10 * time.Second,
			Clock:               &systemClock{},
		})

		if err != nil {
			log.WithFields(log.Fields{"customer_id": account.CustomerID}).WithError(err).Error("error fetching credentials from AWS")
		}

		return creds, err
	} else {
		return creds, AccountNotFound
	}
}

type systemClock struct{}

func (s *systemClock) Now() time.Time {
	return time.Now()
}

func handleAWSError(meth string, err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		log.WithError(err).Errorf("IAM error - %s", meth)

		switch awsErr.Code() {
		case "AccessDenied":
			return InsufficientPermissions

		case "EntityAlreadyExists": // ignoring this
			return nil
		}
	}

	return err
}

func parseARNAccount(arn string) (int, error) {
	matches := arnRegexp.FindStringSubmatch(arn)
	if len(matches) < 2 {
		return 0, fmt.Errorf("No account ID match in ARN: %s", arn)
	}

	if matches[1] == "" {
		return 0, fmt.Errorf("No account ID match in ARN: %s", arn)
	}

	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return id, nil
}

func getEnvRegion() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	if region == "" {
		// not sure it matters really
		region = "us-west-1"
	}

	return region
}
