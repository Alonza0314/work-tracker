package processor

import (
	"archive/zip"
	"backend/constant"
	"backend/model"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func mustBackup(t *testing.T, p *Processor) []byte {
	t.Helper()

	data, fileName, errDetail := p.CreateBackup()
	if errDetail != nil {
		t.Fatalf("CreateBackup: %+v", errDetail)
	}
	if !strings.HasPrefix(fileName, "work-tracker-backup_") || !strings.HasSuffix(fileName, ".zip") {
		t.Errorf("file name = %q", fileName)
	}
	return data
}

func zipFiles(t *testing.T, data []byte) map[string][]byte {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("not a zip: %v", err)
	}
	files := map[string][]byte{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, _ := io.ReadAll(rc)
		_ = rc.Close()
		files[file.Name] = content
	}
	return files
}

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(content))
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBackupZipHasManifestAndOneJsonPerTable(t *testing.T) {
	f := newWorkFixture(t)
	f.mustCreateRecord(t, f.alice, "2026-09-01", 2)

	files := zipFiles(t, mustBackup(t, f.p))

	for _, name := range []string{"manifest.json", "account.json", "category.json", "project.json", "work.json", "todo.json", "holiday.json", "setting.json"} {
		if _, ok := files[name]; !ok {
			t.Errorf("zip is missing %s", name)
		}
	}
	manifest := model.BackupManifest{}
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil || manifest.App != "work-tracker" || manifest.Version != 1 {
		t.Errorf("manifest = %+v, %v", manifest, err)
	}
	var records []model.WorkRecord
	if err := json.Unmarshal(files["work.json"], &records); err != nil || len(records) != 1 || records[0].Account != "ALICE" {
		t.Errorf("work.json = %s", files["work.json"])
	}
}

func TestRestoreBackupReplacesAllData(t *testing.T) {
	source := newWorkFixture(t)
	source.mustCreateRecord(t, source.alice, "2026-09-01", 2)
	data := mustBackup(t, source.p)

	target := newTestProcessorAt(t, filepath.Join(t.TempDir(), "target.db"), "root", "pw")
	mustCreateUser(t, target, "bob", constant.ROLE_DEFAULT)

	if _, errDetail := target.RestoreBackup(data); errDetail != nil {
		t.Fatalf("RestoreBackup: %+v", errDetail)
	}

	if _, err := target.GetAccount("BOB"); err == nil {
		t.Error("data from before the restore is still there")
	}
	if _, err := target.GetAccount("ALICE"); err != nil {
		t.Errorf("restored account missing: %v", err)
	}
	records, _ := target.ListWorkRecords(&model.WorkRecordFilter{})
	if len(records) != 1 {
		t.Errorf("restored records = %d, want 1", len(records))
	}
	// this server's config admin can still sign in, and is the system admin
	root := mustGet(t, target, "root")
	if !root.IsSystem {
		t.Error("config admin is not the system admin after restore")
	}
	if admin := mustGet(t, target, testAdminAccount); admin.IsSystem {
		t.Error("the backup's system admin kept its system flag")
	}
	if _, errDetail := target.Login(&model.RequestLogin{Account: "root", Password: "pw"}); errDetail != nil {
		t.Errorf("config admin login after restore: %+v", errDetail)
	}
}

func TestRestoreBackupRejectsOtherFiles(t *testing.T) {
	f := newWorkFixture(t)

	cases := map[string][]byte{
		"not a zip":         []byte("hello"),
		"no manifest":       makeZip(t, map[string]string{"account.json": "[]"}),
		"other app":         makeZip(t, map[string]string{"manifest.json": `{"app":"other","version":1}`}),
		"newer version":     makeZip(t, map[string]string{"manifest.json": `{"app":"work-tracker","version":99}`}),
		"broken table":      makeZip(t, map[string]string{"manifest.json": `{"app":"work-tracker","version":1}`, "work.json": "{"}),
		"duplicate account": makeZip(t, map[string]string{"manifest.json": `{"app":"work-tracker","version":1}`, "account.json": `[{"account":"X"},{"account":"X"}]`}),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			_, errDetail := f.p.RestoreBackup(data)
			expectStatus(t, errDetail, http.StatusBadRequest)
			if _, err := f.p.GetAccount("ALICE"); err != nil {
				t.Errorf("a rejected restore changed the data: %v", err)
			}
		})
	}
}

func TestResetRequiresConfirmation(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.ResetAll(&model.RequestReset{Confirm: "reset"})
	expectStatus(t, errDetail, http.StatusBadRequest)
	if _, err := f.p.GetAccount("ALICE"); err != nil {
		t.Errorf("an unconfirmed reset removed data: %v", err)
	}
}

func TestResetLeavesOnlyTheSystemAdmin(t *testing.T) {
	f := newWorkFixture(t)
	f.mustCreateRecord(t, f.alice, "2026-09-01", 2)
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{AllowViewAll: boolPtr(true)}); errDetail != nil {
		t.Fatal(errDetail)
	}

	if _, errDetail := f.p.ResetAll(&model.RequestReset{Confirm: "RESET"}); errDetail != nil {
		t.Fatalf("ResetAll: %+v", errDetail)
	}

	backup, err := f.p.Dump()
	if err != nil {
		t.Fatal(err)
	}
	if len(backup.Accounts) != 1 || !backup.Accounts[0].IsSystem || backup.Accounts[0].Account != "ADMIN" {
		t.Errorf("accounts = %+v", backup.Accounts)
	}
	if len(backup.WorkRecords)+len(backup.Categories)+len(backup.Projects)+len(backup.Todos) != 0 || backup.WorkSetting.AllowViewAll {
		t.Errorf("left after reset = %+v", backup)
	}
	if _, errDetail := f.p.Login(&model.RequestLogin{Account: testAdminAccount, Password: testAdminPassword}); errDetail != nil {
		t.Errorf("system admin login after reset: %+v", errDetail)
	}
}
