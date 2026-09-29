package context

import (
	"backend/model"
	"errors"
	"fmt"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account already exists")
)

// DbIf is the storage abstraction; add a new implementation plus a case in
// newDb to support another database.
type DbIf interface {
	AccountDbIf

	Release() error
}

type AccountDbIf interface {
	GetAccount(account string) (*model.Account, error)
	ListAccounts() ([]*model.Account, error)
	CreateAccount(acc *model.Account) error
	UpdateAccount(acc *model.Account) error
	DeleteAccount(account string) error
}

func newDb(dbType, dbPath string) (DbIf, error) {
	switch dbType {
	case "bbolt":
		return newBboltDb(dbPath)
	}
	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
