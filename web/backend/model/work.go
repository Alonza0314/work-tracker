package model

import "time"

// WorkOption is an admin-managed dropdown item: a task category or a project.
// Inactive options are hidden from new entries but still name old ones.
// Only categories have a Color (a name from constant.WORK_CATEGORY_COLORS).
type WorkOption struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Color  string `json:"color,omitempty"`
}

// WorkEntry holds the fields shared by work records and todos. Optional
// fields are "" / 0 when not set: a record may have no project, a todo may
// also have no category or hours.
type WorkEntry struct {
	Date        string  `json:"date"`
	CategoryID  string  `json:"categoryId"`
	Description string  `json:"description"`
	Hours       float64 `json:"hours"`
	ProjectID   string  `json:"projectId"`
}

type WorkRecord struct {
	ID      string `json:"id"`
	Account string `json:"account"`

	WorkEntry

	CreatedAt time.Time `json:"createdAt"`
}

type Todo struct {
	ID      string `json:"id"`
	Account string `json:"account"`

	WorkEntry

	CreatedAt time.Time `json:"createdAt"`
}

// WorkSetting holds the work table settings. StartDate (YYYY-MM-DD, empty
// when unset) is when the team started logging: missed entries are not
// checked before it.
type WorkSetting struct {
	AllowViewAll bool   `json:"allowViewAll"`
	StartDate    string `json:"startDate,omitempty"`
}

// WorkRecordFilter fields are optional; empty fields match everything.
// From and To are inclusive YYYY-MM-DD dates.
type WorkRecordFilter struct {
	Account    string
	From       string
	To         string
	CategoryID string
	ProjectID  string
}

type WorkMember struct {
	Account string `json:"account"`
	Name    string `json:"name"`
}

type ResponseWorkOptions struct {
	Message      string       `json:"message"`
	Categories   []WorkOption `json:"categories"`
	Projects     []WorkOption `json:"projects"`
	AllowViewAll bool         `json:"allowViewAll"`
	StartDate    string       `json:"startDate,omitempty"`
}

type RequestCreateWorkOption struct {
	Name string `json:"name" binding:"required"`
}

// RequestUpdateWorkOption fields are optional; only non-nil fields are applied.
type RequestUpdateWorkOption struct {
	Name   *string `json:"name" binding:"omitempty,min=1"`
	Active *bool   `json:"active"`
	Color  *string `json:"color"`
}

type ResponseWorkOption struct {
	Message string      `json:"message"`
	Option  *WorkOption `json:"option,omitempty"`
}

// ResponseDeleteWorkOption reports how many work records and todos lost the
// deleted option.
type ResponseDeleteWorkOption struct {
	Message string `json:"message"`
	Cleared int    `json:"cleared"`
}

// RequestUpdateWorkSetting fields are optional; only non-nil fields are
// applied. An empty StartDate clears it.
type RequestUpdateWorkSetting struct {
	AllowViewAll *bool   `json:"allowViewAll"`
	StartDate    *string `json:"startDate"`
}

type ResponseWorkSetting struct {
	Message      string `json:"message"`
	AllowViewAll bool   `json:"allowViewAll"`
	StartDate    string `json:"startDate,omitempty"`
}

// RequestSaveWorkEntry creates or fully replaces a work record or a todo.
// Every field is optional: a record without a date gets today, a todo may
// have none, and unset hours are 0 (see processor.checkWorkEntry).
type RequestSaveWorkEntry struct {
	Date        string  `json:"date"`
	CategoryID  string  `json:"categoryId"`
	Description string  `json:"description"`
	Hours       float64 `json:"hours" binding:"omitempty,gt=0,lte=24"`
	ProjectID   string  `json:"projectId"`
}

// RequestListMyWorkRecords lists the records dated From..To (inclusive
// YYYY-MM-DD); the frontend asks for one week at a time.
type RequestListMyWorkRecords struct {
	From string `form:"from" binding:"required"`
	To   string `form:"to" binding:"required"`
}

// RequestListWorkRecords lists everyone's records dated From..To (inclusive);
// the other fields are optional filters.
type RequestListWorkRecords struct {
	From       string `form:"from" binding:"required"`
	To         string `form:"to" binding:"required"`
	Account    string `form:"account"`
	CategoryID string `form:"categoryId"`
	ProjectID  string `form:"projectId"`
}

type ResponseWorkRecordList struct {
	Message    string       `json:"message"`
	Records    []WorkRecord `json:"records"`
	Total      int          `json:"total"`
	TotalHours float64      `json:"totalHours"`
}

type ResponseWorkRecord struct {
	Message string      `json:"message"`
	Record  *WorkRecord `json:"record,omitempty"`
}

type ResponseDeleteWorkRecord struct {
	Message string `json:"message"`
}

type ResponseTodos struct {
	Message string `json:"message"`
	Todos   []Todo `json:"todos"`
}

type ResponseTodo struct {
	Message string `json:"message"`
	Todo    *Todo  `json:"todo,omitempty"`
}

type ResponseDeleteTodo struct {
	Message string `json:"message"`
}

// RequestCompleteTodo optionally fills (or overrides) the todo's category,
// hours and project. The record keeps the todo's date; Date (the client's
// today) is only used for a todo without one.
type RequestCompleteTodo struct {
	Date       string  `json:"date"`
	CategoryID string  `json:"categoryId"`
	Hours      float64 `json:"hours" binding:"omitempty,gt=0,lte=24"`
	ProjectID  string  `json:"projectId"`
}

type ResponseWorkMembers struct {
	Message string       `json:"message"`
	Members []WorkMember `json:"members"`
}
