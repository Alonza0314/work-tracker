package context

import (
	"backend/model"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrAccountExists      = errors.New("account already exists")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrProjectNotFound    = errors.New("project not found")
	ErrWorkRecordNotFound = errors.New("work record not found")
	ErrTodoNotFound       = errors.New("todo not found")
	ErrHolidayNotFound    = errors.New("holiday not found")
	ErrApiTokenNotFound   = errors.New("api token not found")

	errSettingNotFound = errors.New("setting not found")
)

// DbIf is the storage abstraction; add a new implementation plus a case in
// newDb to support another database.
type DbIf interface {
	AccountDbIf
	CategoryDbIf
	ProjectDbIf
	WorkRecordDbIf
	TodoDbIf
	SettingDbIf
	HolidayDbIf
	ApiTokenDbIf
	BackupDbIf

	Release() error
}

type AccountDbIf interface {
	GetAccount(account string) (*model.Account, error)
	ListAccounts() ([]*model.Account, error)
	CreateAccount(acc *model.Account) error
	UpdateAccount(acc *model.Account) error
	// DeleteAccount also deletes the account's API tokens, in one transaction.
	DeleteAccount(account string) error
	// RenameAccount moves the account to a new key and re-owns its work
	// records, todos and API tokens, in one transaction. It fails with
	// ErrAccountExists when newAccount is taken.
	RenameAccount(oldAccount, newAccount string) error
}

// Create* methods of the work interfaces assign the new ID to their argument.
// IDs sort in creation order. Deleting a category or project clears it from
// every work record and todo in the same transaction and returns how many
// entries were cleared.

type CategoryDbIf interface {
	GetCategory(id string) (*model.WorkOption, error)
	ListCategories() ([]*model.WorkOption, error)
	CreateCategory(category *model.WorkOption) error
	UpdateCategory(category *model.WorkOption) error
	DeleteCategory(id string) (int, error)
}

type ProjectDbIf interface {
	GetProject(id string) (*model.WorkOption, error)
	ListProjects() ([]*model.WorkOption, error)
	CreateProject(project *model.WorkOption) error
	UpdateProject(project *model.WorkOption) error
	DeleteProject(id string) (int, error)
}

type WorkRecordDbIf interface {
	GetWorkRecord(id string) (*model.WorkRecord, error)
	ListWorkRecords(filter *model.WorkRecordFilter) ([]*model.WorkRecord, error)
	CreateWorkRecord(record *model.WorkRecord) error
	UpdateWorkRecord(record *model.WorkRecord) error
	DeleteWorkRecord(id string) error
}

type TodoDbIf interface {
	GetTodo(id string) (*model.Todo, error)
	ListTodos(account string) ([]*model.Todo, error)
	CreateTodo(todo *model.Todo) error
	UpdateTodo(todo *model.Todo) error
	DeleteTodo(id string) error
	// CompleteTodo atomically deletes the todo and creates the work record.
	CompleteTodo(todoID string, record *model.WorkRecord) error
}

type SettingDbIf interface {
	// GetWorkSetting returns the zero value when nothing is stored yet.
	GetWorkSetting() (*model.WorkSetting, error)
	UpdateWorkSetting(setting *model.WorkSetting) error
}

// HolidayDbIf stores calendar exceptions keyed by date and source, so a gov
// and a manual entry can exist for the same date.
type HolidayDbIf interface {
	// ListHolidays returns entries dated from..to (inclusive), by date then
	// source.
	ListHolidays(from, to string) ([]*model.Holiday, error)
	// SaveHoliday creates or replaces the entry of holiday.Date and
	// holiday.Source.
	SaveHoliday(holiday *model.Holiday) error
	DeleteHoliday(date, source string) error
	// ReplaceGovHolidays swaps the year's gov entries for holidays in one
	// transaction and returns how many were stored.
	ReplaceGovHolidays(year string, holidays []*model.Holiday) (int, error)
	// GetHolidaySyncedAt returns the zero time before the first sync.
	GetHolidaySyncedAt() (time.Time, error)
	SetHolidaySyncedAt(at time.Time) error
}

// ApiTokenDbIf stores personal access tokens keyed by the SHA-256 of the
// token (model.ApiToken.Hash); the token itself is never stored.
type ApiTokenDbIf interface {
	GetApiToken(hash string) (*model.ApiToken, error)
	ListApiTokens(account string) ([]*model.ApiToken, error)
	// CreateApiToken assigns the new ID to token.
	CreateApiToken(token *model.ApiToken) error
	DeleteApiToken(id string) error
	// TouchApiToken records when the token was last used.
	TouchApiToken(hash string, at time.Time) error
}

// BackupDbIf moves the whole database in and out as a model.Backup.
type BackupDbIf interface {
	// Dump reads every record from one consistent snapshot.
	Dump() (*model.Backup, error)
	// Restore replaces all data with backup in one transaction; new IDs
	// continue after the restored ones.
	Restore(backup *model.Backup) error
	// Reset deletes all data in one transaction.
	Reset() error
}

func newDb(dbType, dbPath string) (DbIf, error) {
	switch dbType {
	case "bbolt":
		return newBboltDb(dbPath)
	}
	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
