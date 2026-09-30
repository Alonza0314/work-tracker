package integrationtest

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

// The holiday calendar: syncing the government calendar, admin entries that
// win over it, and its effect on the hours a week requires.
func TestHolidays(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	year := time.Now().Year()
	dayOff, makeupDay, _ := calendarDays(year)
	yearPath := "/api/holidays?year=" + strconv.Itoa(year)

	holidaysOf := func(t *testing.T) map[string]map[string]any {
		r := alice.get(yearPath)
		expect(t, r, http.StatusOK, "list holidays")
		byDate := map[string]map[string]any{}
		for _, h := range r.list("holidays") {
			byDate[h["date"].(string)] = h
		}
		return byDate
	}

	t.Run("sync stores weekday holidays and weekend workdays", func(t *testing.T) {
		startCalendar(t, http.StatusOK)

		r := admin.post("/api/holidays/sync", nil)
		expect(t, r, http.StatusOK, "sync")
		// two entries for this year and two for the next
		if r.num("synced") != 4 || r.str("lastSyncedAt") == "" {
			t.Errorf("sync = %s", r.body)
		}

		holidays := holidaysOf(t)
		if len(holidays) != 2 {
			t.Fatalf("holidays = %v", holidays)
		}
		if h := holidays[dayOff.Format("2006-01-02")]; h["type"] != "holiday" || h["source"] != "gov" || h["name"] != "測試假日" {
			t.Errorf("day off = %v", h)
		}
		if h := holidays[makeupDay.Format("2006-01-02")]; h["type"] != "workday" || h["name"] == "" {
			t.Errorf("makeup day = %v", h)
		}
		if alice.get(yearPath).str("lastSyncedAt") == "" {
			t.Error("lastSyncedAt missing from the list")
		}
	})

	t.Run("a failed sync keeps the calendar", func(t *testing.T) {
		startCalendar(t, http.StatusInternalServerError)
		expect(t, admin.post("/api/holidays/sync", nil), http.StatusBadGateway, "sync from a broken calendar")
		if len(holidaysOf(t)) != 2 {
			t.Error("a failed sync changed the calendar")
		}
	})

	t.Run("the week of a synced day off needs 32 hours", func(t *testing.T) {
		monday := dayOff.Format("2006-01-02")
		r := alice.get("/api/me/week-summary?from=" + monday)
		expect(t, r, http.StatusOK, "week summary")
		if r.num("requiredHours") != 32 || r.num("daysOff") != 1 {
			t.Errorf("summary = %s", r.body)
		}
	})

	t.Run("an admin entry wins over the government one", func(t *testing.T) {
		date := dayOff.Format("2006-01-02")
		r := admin.put("/api/holidays/"+date, map[string]string{"name": "We work", "type": "workday"})
		expect(t, r, http.StatusOK, "save manual entry")
		if h := holidaysOf(t)[date]; h["source"] != "manual" || h["type"] != "workday" {
			t.Errorf("entry = %v", h)
		}
		if alice.get("/api/me/week-summary?from="+date).num("requiredHours") != 40 {
			t.Error("the manual workday did not restore 40 hours")
		}

		expect(t, admin.del("/api/holidays/"+date), http.StatusOK, "delete manual entry")
		if h := holidaysOf(t)[date]; h["source"] != "gov" {
			t.Errorf("after delete = %v", h)
		}
		expect(t, admin.del("/api/holidays/"+date), http.StatusNotFound, "delete the gov entry")
	})

	t.Run("manual entry validation", func(t *testing.T) {
		expect(t, admin.put("/api/holidays/2026-13-01", map[string]string{"name": "x", "type": "holiday"}), http.StatusBadRequest, "bad date")
		expect(t, admin.put("/api/holidays/2026-12-31", map[string]string{"name": " ", "type": "holiday"}), http.StatusBadRequest, "blank name")
		expect(t, admin.put("/api/holidays/2026-12-31", map[string]string{"name": "x", "type": "party"}), http.StatusBadRequest, "bad type")
		expect(t, alice.get("/api/holidays?year=26"), http.StatusBadRequest, "bad year")
	})

	t.Run("only admins change the calendar", func(t *testing.T) {
		expect(t, alice.post("/api/holidays/sync", nil), http.StatusForbidden, "sync as default user")
		expect(t, alice.put("/api/holidays/2026-12-31", map[string]string{"name": "x", "type": "holiday"}), http.StatusForbidden, "save as default user")
	})
}
