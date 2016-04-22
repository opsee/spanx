package store

import (
	"time"

	"github.com/opsee/basic/com"
)

type Store interface {
	PutAccount(*com.Account) error
	UpdateAccount(*com.Account) error
	ReplaceAccount(oldAccount, newAccount *com.Account) error
	DeleteAccount(*com.Account) error

	GetStack(customerID, externalID string) (*Stack, error)
	PutStack(*Stack) error
	UpdateStack(*Stack, *Stack) error
	DeleteStack(*Stack) error

	GetAccount(*GetAccountRequest) (*com.Account, error)
	GetAccountByExternalID(string) (*com.Account, error)
	GetAccountByCustomerID(string) (*com.Account, error)
}

type GetAccountRequest struct {
	CustomerID string
	Active     bool
}

type Stack struct {
	ExternalID string    `json:"external_id" db:"external_id"`
	CustomerID string    `json:"customer_id" db:"customer_id"`
	StackID    string    `json:"stack_id" db:"stack_id"`
	StackName  string    `json:"stack_name" db:"stack_name"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
