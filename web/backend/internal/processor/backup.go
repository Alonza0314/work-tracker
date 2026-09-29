package processor

import (
	"archive/zip"
	"backend/constant"
	"backend/model"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const (
	backupManifestFile = "manifest.json"
	// a single table file larger than this is refused (zip bomb guard)
	backupMaxFileBytes = 200 << 20
)

// backupSettings is setting.json: the settings bucket's values.
type backupSettings struct {
	Work            model.WorkSetting `json:"work"`
	HolidaySyncedAt time.Time         `json:"holidaySyncedAt"`
}

// backupTables maps each zip entry to its part of a model.Backup.
func backupTables(backup *model.Backup, settings *backupSettings) map[string]any {
	return map[string]any{
		"account.json":  &backup.Accounts,
		"category.json": &backup.Categories,
		"project.json":  &backup.Projects,
		"work.json":     &backup.WorkRecords,
		"todo.json":     &backup.Todos,
		"holiday.json":  &backup.Holidays,
		"setting.json":  settings,
	}
}

// CreateBackup zips a manifest and one JSON file per table. The accounts
// include their password hashes, so a restore can sign everyone in.
func (p *Processor) CreateBackup() ([]byte, string, *model.ErrorDetail) {
	p.ProcLog.Infoln("Creating a backup")

	backup, err := p.Dump()
	if err != nil {
		p.ProcLog.Errorf("Failed to dump the database: %v", err)
		return nil, "", errInternal("Failed to read the database")
	}
	settings := &backupSettings{Work: backup.WorkSetting, HolidaySyncedAt: backup.HolidaySyncedAt}
	now := p.now()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	write := func(name string, value any) error {
		file, err := writer.Create(name)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}

	if err := write(backupManifestFile, model.BackupManifest{
		App:       constant.BACKUP_APP,
		Version:   constant.BACKUP_VERSION,
		CreatedAt: now,
	}); err != nil {
		p.ProcLog.Errorf("Failed to write the backup: %v", err)
		return nil, "", errInternal("Failed to write the backup")
	}
	for name, value := range backupTables(backup, settings) {
		if err := write(name, value); err != nil {
			p.ProcLog.Errorf("Failed to write %s to the backup: %v", name, err)
			return nil, "", errInternal("Failed to write the backup")
		}
	}
	if err := writer.Close(); err != nil {
		p.ProcLog.Errorf("Failed to write the backup: %v", err)
		return nil, "", errInternal("Failed to write the backup")
	}

	return buf.Bytes(), fmt.Sprintf("work-tracker-backup_%s.zip", now.Format("20060102-150405")), nil
}

// RestoreBackup replaces all data with a backup zip made by CreateBackup. A
// file that is not such a backup is refused without touching any data. The
// config admin is re-applied afterwards, so it can always sign in.
func (p *Processor) RestoreBackup(data []byte) (*model.ResponseRestore, *model.ErrorDetail) {
	p.ProcLog.Infoln("Restoring a backup")

	backup, err := readBackup(data)
	if err != nil {
		p.ProcLog.Warnf("Refused a backup: %v", err)
		return nil, errBadRequest("Not a valid backup: " + err.Error())
	}

	if err := p.Restore(backup); err != nil {
		p.ProcLog.Errorf("Failed to restore the backup: %v", err)
		return nil, errInternal("Failed to restore the backup")
	}
	if err := p.InitSystemAdmin(); err != nil {
		p.ProcLog.Errorf("Failed to init the system admin after restore: %v", err)
		return nil, errInternal("Failed to init the system admin")
	}
	p.ProcLog.Infof("Backup restored: %d accounts, %d work records", len(backup.Accounts), len(backup.WorkRecords))

	return &model.ResponseRestore{
		Message: "Restore successful",
	}, nil
}

func readBackup(data []byte) (*model.Backup, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a zip file")
	}
	files := map[string]*zip.File{}
	for _, file := range reader.File {
		files[file.Name] = file
	}

	manifest := model.BackupManifest{}
	if err := readBackupFile(files, backupManifestFile, &manifest, true); err != nil {
		return nil, err
	}
	if manifest.App != constant.BACKUP_APP {
		return nil, fmt.Errorf("not a Work Tracker backup")
	}
	if manifest.Version < 1 || manifest.Version > constant.BACKUP_VERSION {
		return nil, fmt.Errorf("unsupported backup version %d", manifest.Version)
	}

	backup := &model.Backup{}
	settings := &backupSettings{}
	for name, value := range backupTables(backup, settings) {
		if err := readBackupFile(files, name, value, false); err != nil {
			return nil, err
		}
	}
	backup.WorkSetting = settings.Work
	backup.HolidaySyncedAt = settings.HolidaySyncedAt

	seen := map[string]bool{}
	for _, acc := range backup.Accounts {
		if acc.Account == "" || seen[acc.Account] {
			return nil, fmt.Errorf("account.json has a blank or duplicate account %q", acc.Account)
		}
		seen[acc.Account] = true
	}
	return backup, nil
}

// readBackupFile decodes one zip entry into value; a missing optional entry
// leaves value empty.
func readBackupFile(files map[string]*zip.File, name string, value any, required bool) error {
	file, ok := files[name]
	if !ok {
		if required {
			return fmt.Errorf("%s is missing", name)
		}
		return nil
	}
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("cannot read %s", name)
	}
	defer func() { _ = rc.Close() }()

	content, err := io.ReadAll(io.LimitReader(rc, backupMaxFileBytes+1))
	if err != nil {
		return fmt.Errorf("cannot read %s", name)
	}
	if len(content) > backupMaxFileBytes {
		return fmt.Errorf("%s is too large", name)
	}
	if err := json.Unmarshal(content, value); err != nil {
		return fmt.Errorf("%s is not valid JSON", name)
	}
	return nil
}

// ResetAll deletes all data, like a fresh install: only the config admin is
// left, and the government holidays are synced again in the background.
func (p *Processor) ResetAll(req *model.RequestReset) (*model.ResponseReset, *model.ErrorDetail) {
	if req.Confirm != constant.RESET_CONFIRM {
		return nil, errBadRequest("Confirm must be " + constant.RESET_CONFIRM)
	}
	p.ProcLog.Warnf("Resetting all data")

	if err := p.Reset(); err != nil {
		p.ProcLog.Errorf("Failed to reset the database: %v", err)
		return nil, errInternal("Failed to reset the database")
	}
	if err := p.InitSystemAdmin(); err != nil {
		p.ProcLog.Errorf("Failed to init the system admin after reset: %v", err)
		return nil, errInternal("Failed to init the system admin")
	}
	if p.holidaySync && p.holidaySourceUrl != "" {
		go func() {
			if synced, _, err := p.syncHolidays(); err != nil {
				p.ProcLog.Warnf("Failed to sync holidays after reset: %v", err)
			} else {
				p.ProcLog.Infof("Synced %d holidays after reset", synced)
			}
		}()
	}

	return &model.ResponseReset{
		Message: "Reset successful",
	}, nil
}
