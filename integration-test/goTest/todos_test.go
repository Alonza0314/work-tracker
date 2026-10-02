package integrationtest

import (
	"net/http"
	"testing"
)

// Todos: looser required fields, and completing one turns it into a work
// record.
func TestTodos(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	bob := createUser(t, admin, "bob", "Bob", "default")
	category := createOption(t, admin, "categories", "Develop")

	listTodos := func(t *testing.T) []map[string]any {
		r := alice.get("/api/me/todos")
		expect(t, r, http.StatusOK, "list todos")
		return r.list("todos")
	}

	t.Run("nothing is required, not even a date", func(t *testing.T) {
		r := alice.post("/api/me/todos", entry("2026-10-02", "", 0, "write the report"))
		expect(t, r, http.StatusOK, "minimal todo")
		if todo := r.obj("todo"); todo["categoryId"] != "" || todo["hours"] != float64(0) {
			t.Errorf("todo = %v", todo)
		}
		blank := alice.post("/api/me/todos", entry("2026-10-02", "", 0, " "))
		expect(t, blank, http.StatusOK, "blank description")
		if todo := blank.obj("todo"); todo["description"] != "" {
			t.Errorf("todo = %v", todo)
		} else {
			expect(t, alice.del("/api/me/todos/"+todo["id"].(string)), http.StatusOK, "delete blank todo")
		}
		undated := alice.post("/api/me/todos", entry("", "", 0, "some day"))
		expect(t, undated, http.StatusOK, "no date")
		if todo := undated.obj("todo"); todo["date"] != "" {
			t.Errorf("todo = %v", todo)
		} else {
			expect(t, alice.del("/api/me/todos/"+todo["id"].(string)), http.StatusOK, "delete undated todo")
		}
		expect(t, alice.post("/api/me/todos", entry("2026-10-02", "", 1.3, "x")), http.StatusBadRequest, "not half hours")
	})

	t.Run("todos are listed by date", func(t *testing.T) {
		expect(t, alice.post("/api/me/todos", entry("2026-10-01", category, 2, "review")), http.StatusOK, "second todo")
		if got := ids(listTodos(t), "date"); got != "2026-10-01,2026-10-02" {
			t.Errorf("dates = %s", got)
		}
	})

	t.Run("a completed todo keeps its date", func(t *testing.T) {
		todo := listTodos(t)[0] // review on 2026-10-01, has category and hours
		r := alice.post("/api/me/todos/"+todo["id"].(string)+"/complete", map[string]string{"date": "2026-10-05"})
		expect(t, r, http.StatusOK, "complete")
		if record := r.obj("record"); record["date"] != "2026-10-01" || record["description"] != "review" || record["hours"] != float64(2) {
			t.Errorf("record = %v", record)
		}
		if len(listTodos(t)) != 1 {
			t.Error("completed todo still listed")
		}
		if alice.get("/api/me/work-records?from=2026-10-01&to=2026-10-01").num("total") != 1 {
			t.Error("completed todo is not a work record")
		}
	})

	t.Run("a todo without category or hours completes as is", func(t *testing.T) {
		todo := listTodos(t)[0] // write the report
		r := alice.post("/api/me/todos/"+todo["id"].(string)+"/complete", map[string]string{"date": "2026-10-05"})
		expect(t, r, http.StatusOK, "complete")
		if record := r.obj("record"); record["date"] != "2026-10-02" || record["categoryId"] != "" || record["hours"] != float64(0) {
			t.Errorf("record = %v", record)
		}
	})

	t.Run("an undated todo completes on the client's today", func(t *testing.T) {
		r := alice.post("/api/me/todos", entry("", "", 0, "some day"))
		expect(t, r, http.StatusOK, "undated todo")
		r = alice.post("/api/me/todos/"+r.obj("todo")["id"].(string)+"/complete", map[string]string{"date": "2026-10-05"})
		expect(t, r, http.StatusOK, "complete")
		if record := r.obj("record"); record["date"] != "2026-10-05" {
			t.Errorf("record = %v", record)
		}
	})

	t.Run("update, delete and privacy", func(t *testing.T) {
		todo := alice.post("/api/me/todos", entry("2026-10-05", "", 0, "draft")).obj("todo")
		id := todo["id"].(string)

		expect(t, alice.put("/api/me/todos/"+id, entry("2026-10-06", category, 1, "final")), http.StatusOK, "update")
		expect(t, bob.put("/api/me/todos/"+id, entry("2026-10-06", "", 0, "x")), http.StatusNotFound, "bob updates")
		expect(t, bob.post("/api/me/todos/"+id+"/complete", map[string]string{"date": "2026-10-06"}), http.StatusNotFound, "bob completes")
		expect(t, bob.del("/api/me/todos/"+id), http.StatusNotFound, "bob deletes")
		expect(t, alice.del("/api/me/todos/"+id), http.StatusOK, "delete")
		expect(t, alice.del("/api/me/todos/"+id), http.StatusNotFound, "delete again")
	})
}
