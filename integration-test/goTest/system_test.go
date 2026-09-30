package integrationtest

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
)

// Download a backup, change the data, restore it; and what a bad file gets.
func TestBackupRestore(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	category := createOption(t, admin, "categories", "Develop")
	createRecord(t, alice, entry("2026-09-30", category, 2, "before the backup"))
	token := alice.post("/api/me/api-tokens", map[string]any{"name": "skill"}).str("token")

	var backup []byte

	t.Run("download a zip with one JSON file per table", func(t *testing.T) {
		r := admin.get("/api/system/backup")
		expect(t, r, http.StatusOK, "download backup")
		if r.header.Get("Content-Type") != "application/zip" || !strings.Contains(r.header.Get("Content-Disposition"), "work-tracker-backup_") {
			t.Errorf("headers = %v", r.header)
		}
		backup = r.body

		reader, err := zip.NewReader(bytes.NewReader(backup), int64(len(backup)))
		if err != nil {
			t.Fatalf("not a zip: %v", err)
		}
		var names []string
		manifest := map[string]any{}
		for _, file := range reader.File {
			names = append(names, file.Name)
			if file.Name == "manifest.json" {
				rc, _ := file.Open()
				_ = json.NewDecoder(rc).Decode(&manifest)
				_ = rc.Close()
			}
		}
		sort.Strings(names)
		want := "account.json,apitoken.json,category.json,holiday.json,manifest.json,project.json,setting.json,todo.json,work.json"
		if strings.Join(names, ",") != want {
			t.Errorf("files = %v", names)
		}
		if manifest["app"] != "work-tracker" || manifest["version"] != float64(1) {
			t.Errorf("manifest = %v", manifest)
		}
	})

	t.Run("only admins back up and restore", func(t *testing.T) {
		expect(t, alice.get("/api/system/backup"), http.StatusForbidden, "backup as default user")
		expect(t, alice.upload("/api/system/restore", "file", "backup.zip", backup), http.StatusForbidden, "restore as default user")
	})

	t.Run("restore brings back the backed-up data", func(t *testing.T) {
		createUser(t, admin, "carol", "Carol", "default")
		expect(t, admin.del("/api/users/alice"), http.StatusOK, "delete alice")

		r := admin.upload("/api/system/restore", "file", "backup.zip", backup)
		expect(t, r, http.StatusOK, "restore")

		admin := asAdmin(t) // sign in again, as the UI does
		if got := ids(admin.get("/api/users").list("users"), "account"); got != "ADMIN,ALICE" {
			t.Errorf("users = %s", got)
		}
		alice := loginAs(t, "alice", "alice")
		if alice.get("/api/me/work-records?from=2026-09-30&to=2026-09-30").num("total") != 1 {
			t.Error("alice's record was not restored")
		}
		expect(t, (&client{t: t, token: token}).get("/api/me"), http.StatusOK, "API token after restore")

		// new IDs continue after the restored ones
		next := createOption(t, admin, "categories", "After restore")
		if next <= category {
			t.Errorf("new ID %s does not follow %s", next, category)
		}
	})

	t.Run("a file that is not a backup is refused", func(t *testing.T) {
		admin := asAdmin(t)
		expect(t, admin.upload("/api/system/restore", "file", "junk.zip", []byte("hello")), http.StatusBadRequest, "not a zip")

		var buf bytes.Buffer
		writer := zip.NewWriter(&buf)
		file, _ := writer.Create("manifest.json")
		_, _ = file.Write([]byte(`{"app":"other","version":1}`))
		_ = writer.Close()
		expect(t, admin.upload("/api/system/restore", "file", "other.zip", buf.Bytes()), http.StatusBadRequest, "another app's zip")

		expect(t, admin.upload("/api/system/restore", "wrong-field", "backup.zip", backup), http.StatusBadRequest, "no file field")

		if got := ids(admin.get("/api/users").list("users"), "account"); got != "ADMIN,ALICE" {
			t.Errorf("a refused restore changed the data: %s", got)
		}
	})
}

// Reset deletes everything but the config admin, and needs "RESET".
func TestReset(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	category := createOption(t, admin, "categories", "Develop")
	createRecord(t, alice, entry("2026-09-30", category, 2, "work"))
	expect(t, admin.put("/api/settings/work", map[string]any{"allowViewAll": true}), http.StatusOK, "setting")

	t.Run("a reset needs the confirmation word", func(t *testing.T) {
		expect(t, admin.post("/api/system/reset", map[string]string{"confirm": "reset"}), http.StatusBadRequest, "wrong confirm")
		expect(t, admin.post("/api/system/reset", map[string]string{}), http.StatusBadRequest, "no confirm")
		expect(t, alice.post("/api/system/reset", map[string]string{"confirm": "RESET"}), http.StatusForbidden, "reset as default user")
		if len(admin.get("/api/users").list("users")) != 2 {
			t.Error("a refused reset removed users")
		}
	})

	t.Run("reset leaves only the config admin", func(t *testing.T) {
		expect(t, admin.post("/api/system/reset", map[string]string{"confirm": "RESET"}), http.StatusOK, "reset")

		expect(t, alice.get("/api/me"), http.StatusUnauthorized, "alice's token after reset")
		admin := asAdmin(t)
		if got := ids(admin.get("/api/users").list("users"), "account"); got != "ADMIN" {
			t.Errorf("users = %s", got)
		}
		options := admin.get("/api/work/options")
		if len(options.list("categories")) != 0 || options.json["allowViewAll"] != false {
			t.Errorf("options = %s", options.body)
		}
		if admin.get("/api/work-records?from=2026-01-01&to=2026-12-31").num("total") != 0 {
			t.Error("records survived the reset")
		}
	})
}
