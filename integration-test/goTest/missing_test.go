package integrationtest

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// Missed entries: workdays without a work record in the 30 days up to "to".
// Accounts created today skip every earlier day, so the checked days are the
// workdays from today to "to" (two weeks ahead).
func TestMissingEntries(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	createUser(t, admin, "bob", "Bob", "admin")
	category := createOption(t, admin, "categories", "Develop")

	to := date(13)
	var workdays []string
	for day := time.Now(); day.Format("2006-01-02") <= to; day = day.AddDate(0, 0, 1) {
		if day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			workdays = append(workdays, day.Format("2006-01-02"))
		}
	}

	// alice logs the first two workdays; a todo on the third does not count
	createRecord(t, alice, entry(workdays[0], category, 8, "day one"))
	createRecord(t, alice, entry(workdays[1], category, 8, "day two"))
	expect(t, alice.post("/api/me/todos", entry(workdays[2], "", 0, "only a todo")), http.StatusOK, "todo")

	missingOf := func(r *response, account string) map[string]any {
		for _, m := range r.list("members") {
			if m["account"] == account {
				return m
			}
		}
		return nil
	}

	t.Run("workdays without a record are listed, most missing first", func(t *testing.T) {
		r := admin.get("/api/work/missing?to=" + to)
		expect(t, r, http.StatusOK, "missing entries")
		if r.str("to") != to || r.num("checkedCount") != 2 {
			t.Errorf("response = %s", r.body)
		}
		if got := ids(r.list("members"), "account"); got != "BOB,ALICE" {
			t.Errorf("members = %s", got)
		}

		bob := missingOf(r, "BOB")
		if bob["missingCount"] != float64(len(workdays)) {
			t.Errorf("bob missed %v, want %d", bob["missingCount"], len(workdays))
		}
		alice := missingOf(r, "ALICE")
		dates, _ := alice["missingDates"].([]any)
		if alice["missingCount"] != float64(len(workdays)-2) || len(dates) == 0 || dates[0] != workdays[2] {
			t.Errorf("alice = %v, want %d missing from %s", alice, len(workdays)-2, workdays[2])
		}
		if missingOf(r, "ADMIN") != nil {
			t.Error("the system admin is checked")
		}
	})

	t.Run("a day off is not a missed day", func(t *testing.T) {
		expect(t, admin.put("/api/holidays/"+workdays[3], map[string]string{"name": "Day off", "type": "holiday"}), http.StatusOK, "day off")
		r := admin.get("/api/work/missing?to=" + to)
		if missingOf(r, "BOB")["missingCount"] != float64(len(workdays)-1) {
			t.Errorf("bob = %v", missingOf(r, "BOB"))
		}
		expect(t, admin.del("/api/holidays/"+workdays[3]), http.StatusOK, "remove the day off")
	})

	t.Run("checking starts no earlier than the work start date", func(t *testing.T) {
		start := date(7)
		expect(t, admin.put("/api/settings/work", map[string]any{"startDate": start}), http.StatusOK, "start date")

		r := admin.get("/api/work/missing?to=" + to)
		if r.str("from") != start {
			t.Errorf("from = %s, want %s", r.str("from"), start)
		}
		for _, d := range missingOf(r, "BOB")["missingDates"].([]any) {
			if d.(string) < start {
				t.Errorf("checked %s before the start date", d)
			}
		}
	})

	t.Run("default users need the admin's permission", func(t *testing.T) {
		expect(t, alice.get("/api/work/missing?to="+to), http.StatusForbidden, "as default user")
		expect(t, admin.put("/api/settings/work", map[string]any{"allowViewAll": true}), http.StatusOK, "allow")
		expect(t, alice.get("/api/work/missing?to="+to), http.StatusOK, "as default user, allowed")
	})

	t.Run("a bad date is 400", func(t *testing.T) {
		r := admin.get("/api/work/missing?to=" + strings.ReplaceAll(to, "-", "/"))
		expect(t, r, http.StatusBadRequest, "bad to")
	})
}
