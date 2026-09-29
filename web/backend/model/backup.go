package model

import "time"

// Backup is every stored record, independent of the database type.
type Backup struct {
	Accounts        []Account    `json:"accounts"`
	Categories      []WorkOption `json:"categories"`
	Projects        []WorkOption `json:"projects"`
	WorkRecords     []WorkRecord `json:"workRecords"`
	Todos           []Todo       `json:"todos"`
	Holidays        []Holiday    `json:"holidays"`
	WorkSetting     WorkSetting  `json:"workSetting"`
	HolidaySyncedAt time.Time    `json:"holidaySyncedAt"`
}

// BackupManifest identifies a backup zip and its format version.
type BackupManifest struct {
	App       string    `json:"app"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResponseRestore struct {
	Message string `json:"message"`
}

// RequestReset must carry Confirm "RESET", so a stray call wipes nothing.
type RequestReset struct {
	Confirm string `json:"confirm" binding:"required"`
}

type ResponseReset struct {
	Message string `json:"message"`
}
