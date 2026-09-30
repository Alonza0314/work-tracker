package integrationtest

import (
	"net/http"
	"testing"
)

// Task categories, projects and the work table settings on the Work Settings
// page.
func TestWorkOptions(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")

	t.Run("new categories get the least used colors", func(t *testing.T) {
		first := admin.post("/api/categories", map[string]string{"name": "Develop"}).obj("option")
		second := admin.post("/api/categories", map[string]string{"name": "Meeting"}).obj("option")
		if first["color"] != "blue" || second["color"] != "sky" || first["active"] != true {
			t.Errorf("categories = %v, %v", first, second)
		}
		createOption(t, admin, "projects", "Work Tracker")
	})

	t.Run("names are unique ignoring case and not blank", func(t *testing.T) {
		expect(t, admin.post("/api/categories", map[string]string{"name": "develop"}), http.StatusConflict, "duplicate category")
		expect(t, admin.post("/api/projects", map[string]string{"name": " "}), http.StatusBadRequest, "blank project")
	})

	t.Run("rename, recolor and deactivate", func(t *testing.T) {
		id := createOption(t, admin, "categories", "Docs")
		r := admin.put("/api/categories/"+id, map[string]any{"name": "Documents", "color": "purple", "active": false})
		expect(t, r, http.StatusOK, "update category")
		if option := r.obj("option"); option["name"] != "Documents" || option["color"] != "purple" || option["active"] != false {
			t.Errorf("option = %v", option)
		}
		expect(t, admin.put("/api/categories/"+id, map[string]any{"color": "#ff0000"}), http.StatusBadRequest, "unknown color")
		expect(t, admin.put("/api/categories/missing", map[string]any{"active": true}), http.StatusNotFound, "unknown category")

		project := createOption(t, admin, "projects", "Internal")
		expect(t, admin.put("/api/projects/"+project, map[string]any{"color": "blue"}), http.StatusBadRequest, "project color")
	})

	t.Run("everyone reads the options, only admins change them", func(t *testing.T) {
		r := alice.get("/api/work/options")
		expect(t, r, http.StatusOK, "options as default user")
		if len(r.list("categories")) != 3 || len(r.list("projects")) != 2 {
			t.Errorf("options = %s", r.body)
		}
		expect(t, alice.post("/api/categories", map[string]string{"name": "x"}), http.StatusForbidden, "create as default user")
		expect(t, alice.put("/api/settings/work", map[string]any{"allowViewAll": true}), http.StatusForbidden, "settings as default user")
	})

	t.Run("deleting a category or project clears it from records and todos", func(t *testing.T) {
		category := createOption(t, admin, "categories", "Temporary")
		project := createOption(t, admin, "projects", "Temporary")
		record := entry("2026-09-01", category, 1, "work")
		record["projectId"] = project
		created := createRecord(t, alice, record)
		todo := entry("2026-09-02", category, 0, "todo")
		expect(t, alice.post("/api/me/todos", todo), http.StatusOK, "create todo")

		r := admin.del("/api/categories/" + category)
		expect(t, r, http.StatusOK, "delete category")
		if r.num("cleared") != 2 {
			t.Errorf("cleared = %v, want 2", r.num("cleared"))
		}
		expect(t, admin.del("/api/projects/"+project), http.StatusOK, "delete project")

		records := alice.get("/api/me/work-records?from=2026-09-01&to=2026-09-01").list("records")
		if len(records) != 1 || records[0]["id"] != created["id"] || records[0]["categoryId"] != "" || records[0]["projectId"] != "" {
			t.Errorf("records = %v", records)
		}
		expect(t, admin.del("/api/categories/"+category), http.StatusNotFound, "delete again")
	})

	t.Run("work settings update only the given fields", func(t *testing.T) {
		expect(t, admin.put("/api/settings/work", map[string]any{"allowViewAll": true}), http.StatusOK, "allow view all")
		r := admin.put("/api/settings/work", map[string]any{"startDate": "2026-09-01"})
		expect(t, r, http.StatusOK, "start date")
		if r.json["allowViewAll"] != true || r.str("startDate") != "2026-09-01" {
			t.Errorf("setting = %s", r.body)
		}
		options := alice.get("/api/work/options")
		if options.json["allowViewAll"] != true || options.str("startDate") != "2026-09-01" {
			t.Errorf("options = %s", options.body)
		}

		expect(t, admin.put("/api/settings/work", map[string]any{"startDate": "09/01"}), http.StatusBadRequest, "bad start date")
		r = admin.put("/api/settings/work", map[string]any{"startDate": ""})
		if r.str("startDate") != "" {
			t.Errorf("start date not cleared: %s", r.body)
		}
	})
}
