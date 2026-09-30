package integrationtest

import (
	"net/http"
	"testing"
)

// Everyone's work table and its members list: admins always, other users only
// when the admin allows it; filters and totals.
func TestEveryonesWork(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	bob := createUser(t, admin, "bob", "Bob", "default")
	develop := createOption(t, admin, "categories", "Develop")
	meeting := createOption(t, admin, "categories", "Meeting")
	project := createOption(t, admin, "projects", "Work Tracker")

	createRecord(t, alice, entry("2026-09-01", develop, 2, "alice develops"))
	withProject := entry("2026-09-02", meeting, 1.5, "alice meets")
	withProject["projectId"] = project
	createRecord(t, alice, withProject)
	createRecord(t, bob, entry("2026-09-03", develop, 4, "bob develops"))
	createRecord(t, bob, entry("2026-10-01", develop, 8, "next month"))

	const september = "/api/work-records?from=2026-09-01&to=2026-09-30"

	t.Run("default users need the admin's permission", func(t *testing.T) {
		expect(t, alice.get(september), http.StatusForbidden, "records")
		expect(t, alice.get("/api/work/members"), http.StatusForbidden, "members")
		expect(t, alice.get("/api/work/missing?to=2026-09-30"), http.StatusForbidden, "missing entries")
	})

	t.Run("admins see everyone's records with totals", func(t *testing.T) {
		r := admin.get(september)
		expect(t, r, http.StatusOK, "records")
		if r.num("total") != 3 || r.num("totalHours") != 7.5 {
			t.Errorf("total = %v, hours = %v", r.num("total"), r.num("totalHours"))
		}
		if got := ids(r.list("records"), "account"); got != "BOB,ALICE,ALICE" {
			t.Errorf("accounts = %s", got)
		}
	})

	t.Run("filters by member, category and project", func(t *testing.T) {
		cases := map[string]struct {
			query string
			total float64
			hours float64
		}{
			"member (any case)":   {"&account=alice", 2, 3.5},
			"category":            {"&categoryId=" + develop, 2, 6},
			"project":             {"&projectId=" + project, 1, 1.5},
			"member and category": {"&account=BOB&categoryId=" + meeting, 0, 0},
		}
		for name, c := range cases {
			r := admin.get(september + c.query)
			expect(t, r, http.StatusOK, name)
			if r.num("total") != c.total || r.num("totalHours") != c.hours {
				t.Errorf("%s: total = %v, hours = %v", name, r.num("total"), r.num("totalHours"))
			}
		}
	})

	t.Run("a quarter range", func(t *testing.T) {
		r := admin.get("/api/work-records?from=2026-07-01&to=2026-09-30")
		if r.num("total") != 3 {
			t.Errorf("Q3 total = %v", r.num("total"))
		}
		if admin.get("/api/work-records?from=2026-10-01&to=2026-12-31").num("total") != 1 {
			t.Error("Q4 total is not 1")
		}
	})

	t.Run("members list for the filter", func(t *testing.T) {
		members := admin.get("/api/work/members").list("members")
		if got := ids(members, "account"); got != "ADMIN,ALICE,BOB" {
			t.Errorf("members = %s", got)
		}
		if members[1]["name"] != "Alice" {
			t.Errorf("alice = %v", members[1])
		}
	})

	t.Run("allowing everyone opens it to default users", func(t *testing.T) {
		expect(t, admin.put("/api/settings/work", map[string]any{"allowViewAll": true}), http.StatusOK, "allow")
		expect(t, alice.get(september), http.StatusOK, "records")
		expect(t, alice.get("/api/work/members"), http.StatusOK, "members")

		expect(t, admin.put("/api/settings/work", map[string]any{"allowViewAll": false}), http.StatusOK, "disallow")
		expect(t, alice.get(september), http.StatusForbidden, "records after disallowing")
	})
}
