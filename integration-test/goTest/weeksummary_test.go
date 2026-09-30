package integrationtest

import (
	"net/http"
	"testing"
)

// The home page's week: 8 hours per workday, days off taken out and makeup
// days added. 2030-01-07 is a Monday with no government data.
func TestWeekSummary(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	bob := createUser(t, admin, "bob", "Bob", "default")
	category := createOption(t, admin, "categories", "Develop")

	createRecord(t, alice, entry("2030-01-07", category, 8, "monday"))
	createRecord(t, alice, entry("2030-01-08", category, 3.5, "tuesday"))
	createRecord(t, alice, entry("2030-01-12", category, 2, "saturday"))
	createRecord(t, alice, entry("2030-01-14", category, 8, "next week"))
	createRecord(t, bob, entry("2030-01-07", category, 5, "bob"))

	const week = "/api/me/week-summary?from=2030-01-07"

	t.Run("a plain week needs 40 hours", func(t *testing.T) {
		r := alice.get(week)
		expect(t, r, http.StatusOK, "week summary")
		if r.str("to") != "2030-01-13" || len(r.list("days")) != 7 {
			t.Fatalf("summary = %s", r.body)
		}
		if r.num("recordCount") != 3 || r.num("loggedHours") != 13.5 || r.num("requiredHours") != 40 || r.num("remainingHours") != 26.5 {
			t.Errorf("summary = %s", r.body)
		}
		days := r.list("days")
		if days[0]["loggedHours"] != float64(8) || days[0]["requiredHours"] != float64(8) {
			t.Errorf("monday = %v", days[0])
		}
		if days[5]["workday"] != false || days[5]["requiredHours"] != float64(0) || days[5]["loggedHours"] != float64(2) {
			t.Errorf("saturday = %v", days[5])
		}
	})

	t.Run("a Friday off needs 32 hours", func(t *testing.T) {
		expect(t, admin.put("/api/holidays/2030-01-11", map[string]string{"name": "Company trip", "type": "holiday"}), http.StatusOK, "friday off")
		r := alice.get(week)
		if r.num("requiredHours") != 32 || r.num("daysOff") != 1 {
			t.Errorf("summary = %s", r.body)
		}
		if friday := r.list("days")[4]; friday["workday"] != false || friday["holidayName"] != "Company trip" {
			t.Errorf("friday = %v", friday)
		}
	})

	t.Run("a makeup Saturday adds 8 hours", func(t *testing.T) {
		expect(t, admin.put("/api/holidays/2030-01-12", map[string]string{"name": "Makeup", "type": "workday"}), http.StatusOK, "makeup saturday")
		r := alice.get(week)
		if r.num("requiredHours") != 40 {
			t.Errorf("required = %v", r.num("requiredHours"))
		}
		if saturday := r.list("days")[5]; saturday["workday"] != true || saturday["requiredHours"] != float64(8) {
			t.Errorf("saturday = %v", saturday)
		}
	})

	t.Run("remaining hours never go below zero", func(t *testing.T) {
		for _, date := range []string{"2030-01-09", "2030-01-10"} {
			createRecord(t, alice, entry(date, category, 24, "long day"))
		}
		if r := alice.get(week); r.num("remainingHours") != 0 || r.num("loggedHours") != 61.5 {
			t.Errorf("summary = %s", r.body)
		}
	})

	t.Run("a bad date is 400", func(t *testing.T) {
		expect(t, alice.get("/api/me/week-summary?from=01/07"), http.StatusBadRequest, "bad from")
		expect(t, alice.get("/api/me/week-summary"), http.StatusBadRequest, "no from")
	})
}
