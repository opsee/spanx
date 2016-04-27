package store

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) PutAccount(account *com.Account) error {
	return pg.putAccount(pg.db, account)
}

func (pg *Postgres) ReplaceAccount(oldAccount *com.Account, account *com.Account) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	err = pg.deleteAccount(tx, oldAccount)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = pg.putAccount(tx, account)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateAccount only replaces active and role_arn on an account, because
// really those should be the only two mutable fields.
func (pg *Postgres) UpdateAccount(account *com.Account) error {
	_, err := sqlx.NamedExec(
		pg.db,
		`update accounts set active = :active, role_arn = :role_arn where id = :id and customer_id = :customer_id`,
		account,
	)
	return err
}

func (pg *Postgres) DeleteAccount(account *com.Account) error {
	return pg.deleteAccount(pg.db, account)
}

func (pg *Postgres) GetAccountByCustomerID(customerID string) (*com.Account, error) {
	account := &com.Account{}
	err := pg.db.Get(
		account,
		"select * from accounts where customer_id = $1 limit 1",
		customerID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *Postgres) GetAccountByExternalID(externalID string) (*com.Account, error) {
	account := &com.Account{}
	err := pg.db.Get(
		account,
		"select * from accounts where external_id = $1 limit 1",
		externalID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *Postgres) GetAccount(request *GetAccountRequest) (*com.Account, error) {
	account := &com.Account{}
	err := pg.db.Get(
		account,
		"select * from accounts where customer_id = $1 and active = $2 limit 1",
		request.CustomerID,
		request.Active,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *Postgres) putAccount(x sqlx.Ext, account *com.Account) error {
	_, err := sqlx.NamedExec(
		x,
		`insert into accounts (id, customer_id, active) values (:id, :customer_id, :active)`,
		account,
	)
	return err
}

func (pg *Postgres) updateAccount(x sqlx.Ext, account *com.Account) error {
	_, err := sqlx.NamedExec(
		x,
		`update accounts set active = :active where id = :id and customer_id = :customer_id`,
		account,
	)
	return err
}

func (pg *Postgres) deleteAccount(x sqlx.Ext, account *com.Account) error {
	_, err := sqlx.NamedExec(
		x,
		`delete from accounts where id = :id and customer_id = :customer_id`,
		account,
	)
	return err
}

func (pg *Postgres) GetStack(customerID, externalID string) (*Stack, error) {
	stack := &Stack{}
	err := pg.db.Get(
		stack,
		"select * from role_stacks where external_id = $1 and customer_id = $2 limit 1",
		externalID,
		customerID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return stack, nil
}

func (pg *Postgres) PutStack(stack *Stack) error {
	return pg.putStack(pg.db, stack)
}

func (pg *Postgres) UpdateStack(oldStack *Stack, stack *Stack) error {
	tx, err := pg.db.Beginx()
	if err != nil {
		return err
	}

	err = pg.deleteStack(tx, oldStack)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = pg.putStack(tx, stack)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pg *Postgres) DeleteStack(stack *Stack) error {
	return pg.deleteStack(pg.db, stack)
}

func (pg *Postgres) putStack(x sqlx.Ext, stack *Stack) error {
	_, err := sqlx.NamedExec(
		x,
		`insert into role_stacks (external_id, customer_id, stack_id, stack_name, region, active) values (:external_id, :customer_id, :stack_id, :stack_name, :region, :active)`,
		stack,
	)
	return err
}

func (pg *Postgres) deleteStack(x sqlx.Ext, stack *Stack) error {
	_, err := sqlx.NamedExec(
		x,
		`delete from role_stacks where external_id = :external_id and customer_id = :customer_id`,
		stack,
	)
	return err
}
