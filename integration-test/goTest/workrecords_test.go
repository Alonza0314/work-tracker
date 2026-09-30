package integrationtest

import (
	"net/http"
	"testing"
)

// A user's own work records: validation, listing a date range, editing and
// the privacy of other people's records.
func TestWorkRecords(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	bob := createUser(t, admin, "bob", "Bob", "default")
	category := createOption(t, admin, "categories", "Develop")
	project := createOption(t, admin, "projects", "Work Tracker")

	t.Run("validation", func(t *testing.T) {
		cases := map[string]map[string]any{
			"no category":       entry("2026-09-01", "", 1, "x"),
			"no hours":          entry("2026-09-01", category, 0, "x"),
			"not half hours":    entry("2026-09-01", category, 1.3, "x"),
			"too many hours":    entry("2026-09-01", category, 24.5, "x"),
			"bad date":          entry("2026/09/01", category, 1, "x"),
			"blank description": entry("2026-09-01", category, 1, " "),
			"unknown category":  entry("2026-09-01", "missing", 1, "x"),
			"unknown project": func() map[string]any {
				e := entry("2026-09-01", category, 1, "x")
				e["projectId"] = "missing"
				return e
			}(),
		}
		for name, body := range cases {
			expect(t, alice.post("/api/me/work-records", body), http.StatusBadRequest, name)
		}
	})

	t.Run("project is optional, hours in half-hour steps", func(t *testing.T) {
		record := createRecord(t, alice, entry("2026-09-01", category, 1.5, "no project"))
		if record["projectId"] != "" || record["hours"] != 1.5 || record["account"] != "ALICE" {
			t.Errorf("record = %v", record)
		}
	})

	t.Run("multi-line descriptions are kept", func(t *testing.T) {
		record := createRecord(t, alice, entry("2026-09-02", category, 2, "line one\nline two"))
		if record["description"] != "line one\nline two" {
			t.Errorf("description = %q", record["description"])
		}
	})

	t.Run("list a range, newest first, with totals", func(t *testing.T) {
		withProject := entry("2026-09-03", category, 3, "with project")
		withProject["projectId"] = project
		createRecord(t, alice, withProject)
		createRecord(t, bob, entry("2026-09-02", category, 8, "bob's"))

		r := alice.get("/api/me/work-records?from=2026-09-01&to=2026-09-07")
		expect(t, r, http.StatusOK, "list")
		records := r.list("records")
		if got := ids(records, "date"); got != "2026-09-03,2026-09-02,2026-09-01" {
			t.Errorf("dates = %s", got)
		}
		if r.num("total") != 3 || r.num("totalHours") != 6.5 {
			t.Errorf("total = %v, hours = %v", r.num("total"), r.num("totalHours"))
		}

		narrow := alice.get("/api/me/work-records?from=2026-09-02&to=2026-09-02")
		if narrow.num("total") != 1 {
			t.Errorf("one day = %s", narrow.body)
		}
	})

	t.Run("a range is required and must not be reversed", func(t *testing.T) {
		expect(t, alice.get("/api/me/work-records"), http.StatusBadRequest, "no range")
		expect(t, alice.get("/api/me/work-records?from=2026-09-07&to=2026-09-01"), http.StatusBadRequest, "reversed range")
	})

	t.Run("update and delete own records", func(t *testing.T) {
		record := createRecord(t, alice, entry("2026-09-04", category, 1, "draft"))
		id := record["id"].(string)

		updated := alice.put("/api/me/work-records/"+id, entry("2026-09-05", category, 2.5, "final"))
		expect(t, updated, http.StatusOK, "update")
		if r := updated.obj("record"); r["date"] != "2026-09-05" || r["hours"] != 2.5 || r["description"] != "final" {
			t.Errorf("record = %v", r)
		}

		expect(t, alice.del("/api/me/work-records/"+id), http.StatusOK, "delete")
		expect(t, alice.del("/api/me/work-records/"+id), http.StatusNotFound, "delete again")
	})

	t.Run("a record keeps a category deactivated after it was chosen", func(t *testing.T) {
		old := createOption(t, admin, "categories", "Old")
		record := createRecord(t, alice, entry("2026-09-06", old, 1, "old category"))
		expect(t, admin.put("/api/categories/"+old, map[string]any{"active": false}), http.StatusOK, "deactivate")

		expect(t, alice.put("/api/me/work-records/"+record["id"].(string), entry("2026-09-06", old, 2, "still old")),
			http.StatusOK, "update keeping the inactive category")
		expect(t, alice.post("/api/me/work-records", entry("2026-09-06", old, 1, "new")),
			http.StatusBadRequest, "new record with the inactive category")
	})

	t.Run("other people's records answer 404", func(t *testing.T) {
		record := createRecord(t, alice, entry("2026-09-07", category, 1, "alice's"))
		id := record["id"].(string)
		expect(t, bob.put("/api/me/work-records/"+id, entry("2026-09-07", category, 1, "hijack")), http.StatusNotFound, "bob updates")
		expect(t, bob.del("/api/me/work-records/"+id), http.StatusNotFound, "bob deletes")
		expect(t, admin.del("/api/me/work-records/"+id), http.StatusNotFound, "admin deletes via /me")
	})
}
