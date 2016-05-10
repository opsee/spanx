package roler

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/golang-lru"
	"github.com/opsee/basic/com"
	"github.com/opsee/spanx/store"
	"testing"
	"time"
)

func TestParseAccountARN(t *testing.T) {
	acc, err := parseARNAccount("arn:aws:iam::975383256012:root")
	if err != nil {
		t.Fatal(err)
	}

	if acc != 975383256012 {
		t.Fatal("account didn't match")
	}

	acc, err = parseARNAccount("arn:aws:iam::933693344490:user/mark")
	if err != nil {
		t.Fatal(err)
	}

	if acc != 933693344490 {
		t.Fatal("account didn't match")
	}
}

// A very minimal test. I can't test api stuff untilI make a test AWS session.
func TestCache(t *testing.T) {
	db := &testStore{}

	lru, err := lru.New(1)
	if err != nil {
		t.Fatal(err)
	}

	lru.Add(975383256012, &Credentials{
		Expires: time.Now().UTC().Add(90 * time.Second),
		Value: credentials.Value{
			AccessKeyID:     "666",
			SecretAccessKey: "666",
			SessionToken:    "666",
		},
	})

	creds, err := GetCredentials(db, lru, "mycustomerid")
	if err != nil {
		t.Fatal(err)
	}

	if creds.Value.AccessKeyID != "666" {
		t.Error("expected cached credentials, but got: ", creds.Value)
	}
}

func TestGetAccountCredentials(t *testing.T) {
	var (
		err   error
		creds Credentials
	)

	db := &testStore{}
	lru, err := lru.New(1)
	if err != nil {
		t.Fatal(err)
	}

	lru.Add(975383256012, &Credentials{
		Expires: time.Now().UTC().Add(90 * time.Second),
		Value: credentials.Value{
			AccessKeyID:     "666",
			SecretAccessKey: "666",
			SessionToken:    "666",
		},
	})

	_, err = getAccountCredentials(db, lru, nil)
	if err == nil {
		t.Error("getAccountCredentials should not accept nil account: ", err.Error())
	}

	_, err = getAccountCredentials(db, lru, &com.Account{ID: 0})
	if err == nil {
		t.Error("getAccountCredentials should not accept account ID 0: ", err.Error())
	}

	creds, err = getAccountCredentials(db, lru, &com.Account{ID: 975383256012})
	if err != nil {
		t.Error("getAccountCredentials returned an error: ", err.Error())
	}

	if creds.Value.AccessKeyID != "666" {
		t.Error("expected cached credentials, but got: ", creds.Value)
	}
}

type testStore struct{}

func (ts *testStore) PutAccount(_ *com.Account) error {
	return nil
}
func (ts *testStore) UpdateAccount(_ *com.Account) error {
	return nil
}
func (ts *testStore) ReplaceAccount(oldAccount, newAccount *com.Account) error {
	return nil
}
func (ts *testStore) DeleteAccount(_ *com.Account) error {
	return nil
}
func (ts *testStore) GetStack(customerID, externalID string) (*store.Stack, error) {
	return nil, nil
}
func (ts *testStore) PutStack(_ *store.Stack) error {
	return nil
}
func (ts *testStore) UpdateStack(_ *store.Stack, _ *store.Stack) error {
	return nil
}
func (ts *testStore) DeleteStack(_ *store.Stack) error {
	return nil
}
func (ts *testStore) GetAccount(_ *store.GetAccountRequest) (*com.Account, error) {
	return &com.Account{ID: 975383256012}, nil
}
func (ts *testStore) GetAccountByExternalID(_ string) (*com.Account, error) {
	return nil, nil
}
func (ts *testStore) GetAccountByCustomerID(_ string) (*com.Account, error) {
	return nil, nil
}
