package integrationtest

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The fake government office calendar that the app syncs from
// (config/config.yaml: http://host.docker.internal:18889/{year}.json).
const calendarAddr = ":18889"

// calendarDays are the entries the fake calendar publishes for a year:
//   - the first Monday of March is a day off,
//   - the first Saturday of June is a makeup workday,
//   - the first Sunday of March is a holiday on a weekend (ignored by the app).
func calendarDays(year int) (dayOff, makeupDay, weekendHoliday time.Time) {
	return firstWeekday(year, time.March, time.Monday),
		firstWeekday(year, time.June, time.Saturday),
		firstWeekday(year, time.March, time.Sunday)
}

func firstWeekday(year int, month time.Month, weekday time.Weekday) time.Time {
	day := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	for day.Weekday() != weekday {
		day = day.AddDate(0, 0, 1)
	}
	return day
}

// startCalendar serves the fake calendar until the test ends; status is the
// HTTP status to answer with (200 serves the days).
func startCalendar(t *testing.T, status int) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		year, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), ".json"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}

		dayOff, makeupDay, weekendHoliday := calendarDays(year)
		days := []map[string]any{
			{"date": dayOff.Format("20060102"), "week": "一", "isHoliday": true, "description": "測試假日"},
			{"date": makeupDay.Format("20060102"), "week": "六", "isHoliday": false, "description": ""},
			{"date": weekendHoliday.Format("20060102"), "week": "日", "isHoliday": true, "description": "週日假日"},
		}
		_ = json.NewEncoder(w).Encode(days)
	})

	listener, err := net.Listen("tcp", calendarAddr)
	if err != nil {
		t.Fatalf("fake calendar: %v", err)
	}
	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("fake calendar: %v", err)
		}
	}()
	t.Cleanup(func() { _ = server.Close() })
}
