package store

import (
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
)

type Store interface {
	PutAccount(*com.Account) error
	UpdateAccount(*com.Account) error
	ReplaceAccount(oldAccount, newAccount *com.Account) error
	DeleteAccount(*com.Account) error

	GetStack(customerID, externalID string) (*schema.RoleStack, error)
	GetStackByCustomerId(customerID string) (*schema.RoleStack, error)
	PutStack(*schema.RoleStack) error
	UpdateStack(*schema.RoleStack, *schema.RoleStack) error
	DeleteStack(*schema.RoleStack) error

	GetAccount(*GetAccountRequest) (*com.Account, error)
	GetAccountByExternalID(string) (*com.Account, error)
	GetAccountByCustomerID(string) (*com.Account, error)
}

type GetAccountRequest struct {
	CustomerID string
	Active     bool
}
