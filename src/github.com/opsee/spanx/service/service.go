package service

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/opsee/basic/com"
	"github.com/opsee/spanx/roler"
	"github.com/opsee/spanx/store"
)

var (
	errUnauthorized     = errors.New("unauthorized.")
	errMissingAccessKey = errors.New("missing AccessKeyID.")
	errMissingSecretKey = errors.New("missing SecretAccessKey.")
	errMissingRegion    = errors.New("missing region.")
	errUnknown          = errors.New("unknown error.")
)

type ResolveCredentialsRequest struct {
	AccessKeyID     string
	SecretAccessKey string
}

type CredentialsResponse struct {
	Credentials credentials.Value
}

type Service interface {
	ResolveCredentials(*com.User, *ResolveCredentialsRequest) (*CredentialsResponse, error)
	GetCredentials(*com.User) (*CredentialsResponse, error)
}

type service struct {
	db store.Store
}

func New(db store.Store) *service {
	return &service{
		db: db,
	}
}

func (s *service) ResolveCredentials(user *com.User, request *ResolveCredentialsRequest) (*CredentialsResponse, error) {
	creds, err := roler.ResolveCredentials(s.db, user.CustomerID, request.AccessKeyID, request.SecretAccessKey)
	if err != nil {
		return nil, err
	}

	return &CredentialsResponse{creds}, nil
}

func (s *service) GetCredentials(user *com.User) (*CredentialsResponse, error) {
	creds, err := roler.GetCredentials(s.db, user.CustomerID)
	if err != nil {
		return nil, err
	}

	return &CredentialsResponse{creds}, nil
}

func (request *ResolveCredentialsRequest) Validate() error {
	if request.AccessKeyID == "" {
		return errMissingAccessKey
	}

	if request.SecretAccessKey == "" {
		return errMissingSecretKey
	}

	return nil
}
