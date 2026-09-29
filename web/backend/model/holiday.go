package model

import "time"

// Holiday is an exception to the Monday-Friday work week on Date: a day off
// (Type holiday) or a working weekend day (Type workday). A manual entry
// wins over a gov entry on the same date.
type Holiday struct {
	Date   string `json:"date"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

type RequestListHolidays struct {
	Year string `form:"year" binding:"required"`
}

type ResponseHolidays struct {
	Message      string     `json:"message"`
	Holidays     []Holiday  `json:"holidays"`
	LastSyncedAt *time.Time `json:"lastSyncedAt,omitempty"`
}

type RequestSaveHoliday struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=holiday workday"`
}

type ResponseHoliday struct {
	Message string   `json:"message"`
	Holiday *Holiday `json:"holiday,omitempty"`
}

type ResponseDeleteHoliday struct {
	Message string `json:"message"`
}

type ResponseSyncHolidays struct {
	Message      string     `json:"message"`
	Synced       int        `json:"synced"`
	LastSyncedAt *time.Time `json:"lastSyncedAt,omitempty"`
}

// RequestWeekSummary asks for the 7 days starting at From (a Monday in the UI).
type RequestWeekSummary struct {
	From string `form:"from" binding:"required"`
}

type WeekDay struct {
	Date          string  `json:"date"`
	Workday       bool    `json:"workday"`
	HolidayName   string  `json:"holidayName,omitempty"`
	LoggedHours   float64 `json:"loggedHours"`
	RequiredHours float64 `json:"requiredHours"`
}

type ResponseWeekSummary struct {
	Message        string    `json:"message"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	RecordCount    int       `json:"recordCount"`
	LoggedHours    float64   `json:"loggedHours"`
	RequiredHours  float64   `json:"requiredHours"`
	RemainingHours float64   `json:"remainingHours"`
	DaysOff        int       `json:"daysOff"`
	Days           []WeekDay `json:"days"`
}
