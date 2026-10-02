package processor

import (
	"backend/constant"
	"backend/model"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mustSaveHoliday(t *testing.T, f *workFixture, date, name, kind, source string) {
	t.Helper()

	if err := f.p.SaveHoliday(&model.Holiday{Date: date, Name: name, Type: kind, Source: source}); err != nil {
		t.Fatalf("SaveHoliday: %v", err)
	}
}

func mustWeekSummary(t *testing.T, f *workFixture, from string) *model.ResponseWeekSummary {
	t.Helper()

	resp, errDetail := f.p.GetWeekSummary(f.alice, &model.RequestWeekSummary{From: from})
	if errDetail != nil {
		t.Fatalf("GetWeekSummary: %+v", errDetail)
	}
	return resp
}

// week summary; 2026-09-28 is a Monday

func TestWeekSummaryPlainWeek(t *testing.T) {
	f := newWorkFixture(t)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_DEFAULT)
	f.mustCreateRecord(t, f.alice, "2026-09-28", 8)
	f.mustCreateRecord(t, f.alice, "2026-09-29", 3.5)
	f.mustCreateRecord(t, f.alice, "2026-10-03", 2)
	f.mustCreateRecord(t, f.alice, "2026-10-05", 8) // next week
	f.mustCreateRecord(t, bob, "2026-09-28", 5)

	s := mustWeekSummary(t, f, "2026-09-28")

	if s.From != "2026-09-28" || s.To != "2026-10-04" || len(s.Days) != 7 {
		t.Fatalf("range = %s..%s, %d days", s.From, s.To, len(s.Days))
	}
	if s.RecordCount != 3 || s.LoggedHours != 13.5 || s.RequiredHours != 40 || s.RemainingHours != 26.5 || s.DaysOff != 0 {
		t.Errorf("summary = %+v", s)
	}
	if !s.Days[0].Workday || s.Days[0].RequiredHours != 8 || s.Days[0].LoggedHours != 8 {
		t.Errorf("monday = %+v", s.Days[0])
	}
	if s.Days[5].Workday || s.Days[5].RequiredHours != 0 || s.Days[5].LoggedHours != 2 {
		t.Errorf("saturday = %+v", s.Days[5])
	}
}

func TestWeekSummaryHolidaysAndMakeupDays(t *testing.T) {
	f := newWorkFixture(t)
	mustSaveHoliday(t, f, "2026-10-02", "Mid-Autumn", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)
	mustSaveHoliday(t, f, "2026-10-03", "Makeup", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_MANUAL)
	// a holiday on a weekend changes nothing
	mustSaveHoliday(t, f, "2026-10-04", "Sunday holiday", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)

	s := mustWeekSummary(t, f, "2026-09-28")

	if s.RequiredHours != 40 || s.DaysOff != 1 {
		t.Errorf("required = %v, daysOff = %d, want 40 and 1", s.RequiredHours, s.DaysOff)
	}
	friday, saturday := s.Days[4], s.Days[5]
	if friday.Workday || friday.RequiredHours != 0 || friday.HolidayName != "Mid-Autumn" {
		t.Errorf("friday = %+v", friday)
	}
	if !saturday.Workday || saturday.RequiredHours != 8 || saturday.HolidayName != "Makeup" {
		t.Errorf("saturday = %+v", saturday)
	}
}

func TestWeekSummaryFridayOffNeedsFourDays(t *testing.T) {
	f := newWorkFixture(t)
	mustSaveHoliday(t, f, "2026-10-02", "Day off", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)

	if s := mustWeekSummary(t, f, "2026-09-28"); s.RequiredHours != 32 {
		t.Errorf("required = %v, want 32", s.RequiredHours)
	}
}

func TestWeekSummaryManualWinsOverGov(t *testing.T) {
	f := newWorkFixture(t)
	mustSaveHoliday(t, f, "2026-09-30", "Gov holiday", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)
	mustSaveHoliday(t, f, "2026-09-30", "We work", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_MANUAL)

	s := mustWeekSummary(t, f, "2026-09-28")
	if s.RequiredHours != 40 || s.DaysOff != 0 || !s.Days[2].Workday {
		t.Errorf("summary = %+v", s)
	}
}

func TestWeekSummaryRemainingNeverNegative(t *testing.T) {
	f := newWorkFixture(t)
	for _, date := range []string{"2026-09-28", "2026-09-29", "2026-09-30", "2026-10-01", "2026-10-02"} {
		f.mustCreateRecord(t, f.alice, date, 9)
	}

	if s := mustWeekSummary(t, f, "2026-09-28"); s.LoggedHours != 45 || s.RemainingHours != 0 {
		t.Errorf("logged = %v, remaining = %v", s.LoggedHours, s.RemainingHours)
	}
}

func TestWeekSummaryRejectsBadDate(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.GetWeekSummary(f.alice, &model.RequestWeekSummary{From: "09/28"})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

// manual holidays

func TestSaveManualHoliday(t *testing.T) {
	f := newWorkFixture(t)

	resp, errDetail := f.p.SaveManualHoliday("2026-12-31", &model.RequestSaveHoliday{Name: " Year end ", Type: constant.HOLIDAY_TYPE_HOLIDAY})
	if errDetail != nil {
		t.Fatalf("SaveManualHoliday: %+v", errDetail)
	}
	if resp.Holiday.Name != "Year end" || resp.Holiday.Source != constant.HOLIDAY_SOURCE_MANUAL || resp.Holiday.Date != "2026-12-31" {
		t.Errorf("holiday = %+v", resp.Holiday)
	}

	_, errDetail = f.p.SaveManualHoliday("2026/12/31", &model.RequestSaveHoliday{Name: "x", Type: constant.HOLIDAY_TYPE_HOLIDAY})
	expectStatus(t, errDetail, http.StatusBadRequest)
	_, errDetail = f.p.SaveManualHoliday("2026-12-31", &model.RequestSaveHoliday{Name: " ", Type: constant.HOLIDAY_TYPE_HOLIDAY})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestDeleteManualHolidayRevealsGov(t *testing.T) {
	f := newWorkFixture(t)
	mustSaveHoliday(t, f, "2026-10-09", "Gov", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)
	mustSaveHoliday(t, f, "2026-10-09", "Manual", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_MANUAL)

	if _, errDetail := f.p.DeleteManualHoliday("2026-10-09"); errDetail != nil {
		t.Fatalf("DeleteManualHoliday: %+v", errDetail)
	}
	list, _ := f.p.ListYearHolidays(&model.RequestListHolidays{Year: "2026"})
	if len(list.Holidays) != 1 || list.Holidays[0].Name != "Gov" {
		t.Errorf("holidays = %+v", list.Holidays)
	}

	_, errDetail := f.p.DeleteManualHoliday("2026-10-09")
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestListYearHolidaysManualWins(t *testing.T) {
	f := newWorkFixture(t)
	mustSaveHoliday(t, f, "2026-10-10", "Gov", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)
	mustSaveHoliday(t, f, "2026-10-10", "Manual", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_MANUAL)
	mustSaveHoliday(t, f, "2027-01-01", "Next year", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)

	list, errDetail := f.p.ListYearHolidays(&model.RequestListHolidays{Year: "2026"})
	if errDetail != nil {
		t.Fatalf("ListYearHolidays: %+v", errDetail)
	}
	if len(list.Holidays) != 1 || list.Holidays[0].Name != "Manual" {
		t.Errorf("holidays = %+v", list.Holidays)
	}

	_, errDetail = f.p.ListYearHolidays(&model.RequestListHolidays{Year: "26"})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

// sync from the office calendar

const testCalendar2026 = `[
  {"date": "20261009", "week": "五", "isHoliday": true, "description": "國慶日補假"},
  {"date": "20261010", "week": "六", "isHoliday": true, "description": "國慶日"},
  {"date": "20261011", "week": "日", "isHoliday": false, "description": ""},
  {"date": "20261012", "week": "一", "isHoliday": false, "description": ""},
  {"date": "20261013", "week": "二", "isHoliday": true, "description": ""}
]`

func newCalendarServer(t *testing.T, status2026 int) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/2026.json" && status2026 == http.StatusOK {
			_, _ = w.Write([]byte(testCalendar2026))
			return
		}
		if r.URL.Path == "/2026.json" {
			w.WriteHeader(status2026)
			return
		}
		// the next year is not published yet
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestSyncHolidaysStoresWeekdayHolidaysAndWeekendWorkdays(t *testing.T) {
	f := newWorkFixture(t)
	server := newCalendarServer(t, http.StatusOK)
	f.p.holidaySourceUrl = server.URL + "/{year}.json"
	f.p.now = func() time.Time { return time.Date(2026, 9, 29, 9, 0, 0, 0, time.UTC) }
	mustSaveHoliday(t, f, "2026-10-09", "We work", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_MANUAL)

	resp, errDetail := f.p.SyncHolidaysNow()
	if errDetail != nil {
		t.Fatalf("SyncHolidaysNow: %+v", errDetail)
	}
	if resp.Synced != 3 || resp.LastSyncedAt == nil || !resp.LastSyncedAt.Equal(f.p.now()) {
		t.Errorf("sync = %+v", resp)
	}

	list, _ := f.p.ListYearHolidays(&model.RequestListHolidays{Year: "2026"})
	got := map[string]model.Holiday{}
	for _, h := range list.Holidays {
		got[h.Date] = h
	}
	if len(got) != 3 {
		t.Fatalf("holidays = %+v", list.Holidays)
	}
	if got["2026-10-09"].Source != constant.HOLIDAY_SOURCE_MANUAL {
		t.Errorf("manual entry lost: %+v", got["2026-10-09"])
	}
	if h := got["2026-10-11"]; h.Type != constant.HOLIDAY_TYPE_WORKDAY || h.Source != constant.HOLIDAY_SOURCE_GOV || h.Name == "" {
		t.Errorf("makeup workday = %+v", h)
	}
	if h := got["2026-10-13"]; h.Type != constant.HOLIDAY_TYPE_HOLIDAY || h.Name == "" {
		t.Errorf("unnamed holiday = %+v", h)
	}
	if list.LastSyncedAt == nil {
		t.Error("lastSyncedAt missing from the list")
	}
}

func TestSyncHolidaysFailsOnServerError(t *testing.T) {
	f := newWorkFixture(t)
	server := newCalendarServer(t, http.StatusInternalServerError)
	f.p.holidaySourceUrl = server.URL + "/{year}.json"
	f.p.now = func() time.Time { return time.Date(2026, 9, 29, 9, 0, 0, 0, time.UTC) }

	_, errDetail := f.p.SyncHolidaysNow()
	expectStatus(t, errDetail, http.StatusBadGateway)

	list, _ := f.p.ListYearHolidays(&model.RequestListHolidays{Year: "2026"})
	if len(list.Holidays) != 0 || list.LastSyncedAt != nil {
		t.Errorf("failed sync stored data: %+v", list)
	}
}

// missing entries; 2026-09-28 is a Monday, the range is the 30 days to "to"

func mustMissing(t *testing.T, f *workFixture, viewer *model.Account, to string) *model.ResponseMissingEntries {
	t.Helper()

	resp, errDetail := f.p.ListMissingEntries(viewer, &model.RequestMissingEntries{To: to})
	if errDetail != nil {
		t.Fatalf("ListMissingEntries: %+v", errDetail)
	}
	return resp
}

func missingOf(resp *model.ResponseMissingEntries, account string) *model.MissingMember {
	for i := range resp.Members {
		if resp.Members[i].Account == account {
			return &resp.Members[i]
		}
	}
	return nil
}

func TestMissingEntriesCountsWorkdaysWithoutRecords(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)
	// 30 days: 2026-09-01 (Tue) .. 2026-09-30 (Wed) = 22 weekdays
	mustSaveHoliday(t, f, "2026-09-25", "Mid-Autumn", constant.HOLIDAY_TYPE_HOLIDAY, constant.HOLIDAY_SOURCE_GOV)
	mustSaveHoliday(t, f, "2026-09-19", "Makeup", constant.HOLIDAY_TYPE_WORKDAY, constant.HOLIDAY_SOURCE_GOV)
	// alice logs every workday except 09-02 and the makeup Saturday; a todo
	// and a weekend record do not count
	for day := 1; day <= 30; day++ {
		date := fmt.Sprintf("2026-09-%02d", day)
		if date == "2026-09-02" || date == "2026-09-19" {
			continue
		}
		f.mustCreateRecord(t, f.alice, date, 1)
	}
	if _, errDetail := f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-09-02", Description: "todo only"}); errDetail != nil {
		t.Fatal(errDetail)
	}

	resp := mustMissing(t, f, admin, "2026-09-30")

	if resp.From != "2026-09-01" || resp.To != "2026-09-30" {
		t.Errorf("range = %s..%s", resp.From, resp.To)
	}
	// 22 weekdays - 1 holiday + 1 makeup day
	if resp.Workdays != 22 {
		t.Errorf("workdays = %d, want 22", resp.Workdays)
	}
	alice := missingOf(resp, "ALICE")
	if alice == nil || alice.MissingCount != 2 || strings.Join(alice.MissingDates, ",") != "2026-09-02,2026-09-19" || alice.Name != "alice name" {
		t.Errorf("alice = %+v", alice)
	}
}

func TestMissingEntriesMembersAndOrder(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_ADMIN)
	carol := mustCreateUser(t, f.p, "carol", constant.ROLE_DEFAULT)
	// week of 2026-09-28..10-02 (to = Friday, the 30 days start 09-03)
	for _, date := range []string{"2026-09-28", "2026-09-29", "2026-09-30", "2026-10-01", "2026-10-02"} {
		f.mustCreateRecord(t, carol, date, 1)
	}
	f.mustCreateRecord(t, bob, "2026-10-02", 1)

	// work starts on 2026-09-28, so only that week counts
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-09-28")}); errDetail != nil {
		t.Fatal(errDetail)
	}

	resp := mustMissing(t, f, admin, "2026-10-02")

	// the system admin is not checked; admins and default users are
	if resp.CheckedCount != 3 {
		t.Errorf("checked = %d, want 3", resp.CheckedCount)
	}
	if missingOf(resp, "ADMIN") != nil || missingOf(resp, "CAROL") != nil {
		t.Errorf("members = %+v", resp.Members)
	}
	if len(resp.Members) != 2 || resp.Members[0].Account != "ALICE" || resp.Members[0].MissingCount != 5 ||
		resp.Members[1].Account != "BOB" || resp.Members[1].MissingCount != 4 {
		t.Errorf("members = %+v, want ALICE (5) then BOB (4)", resp.Members)
	}
}

func TestMissingEntriesPermission(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.ListMissingEntries(f.alice, &model.RequestMissingEntries{To: "2026-09-30"})
	expectStatus(t, errDetail, http.StatusForbidden)

	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{AllowViewAll: boolPtr(true)}); errDetail != nil {
		t.Fatal(errDetail)
	}
	mustMissing(t, f, f.alice, "2026-09-30")

	_, errDetail = f.p.ListMissingEntries(f.alice, &model.RequestMissingEntries{To: "09/30"})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestCreateUserRecordsCreationTime(t *testing.T) {
	f := newWorkFixture(t)

	if f.alice.CreatedAt.IsZero() {
		t.Error("CreatedAt not set on a new user")
	}
}

// An account created after the work start date is still checked from the
// start date: accounts are often created later than people start working.
func TestMissingEntriesIgnoreAccountCreationTime(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-09-21")}); errDetail != nil {
		t.Fatal(errDetail)
	}
	f.alice.CreatedAt = time.Date(2026, 10, 2, 9, 49, 0, 0, time.Local)
	if err := f.p.UpdateAccount(f.alice); err != nil {
		t.Fatal(err)
	}
	f.mustCreateRecord(t, f.alice, "2026-10-01", 8)

	// 2026-09-21 (Mon) .. 10-01 (Thu): 9 weekdays, one logged
	resp := mustMissing(t, f, admin, "2026-10-01")
	alice := missingOf(resp, "ALICE")
	if alice == nil || alice.MissingCount != 8 || alice.MissingDates[0] != "2026-09-21" {
		t.Errorf("alice = %+v, want 8 missing days from 2026-09-21", alice)
	}
}

func TestMissingEntriesStartAtTheWorkStartDate(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)

	// a start date inside the 30 days moves the range start
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-09-28")}); errDetail != nil {
		t.Fatalf("SaveWorkSetting: %+v", errDetail)
	}
	resp := mustMissing(t, f, admin, "2026-09-30")
	if resp.From != "2026-09-28" || resp.Workdays != 3 || missingOf(resp, "ALICE").MissingCount != 3 {
		t.Errorf("recent start = %s, %d workdays, %+v", resp.From, resp.Workdays, resp.Members)
	}

	// an older start date leaves the 30 days as they are
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-01-01")}); errDetail != nil {
		t.Fatalf("SaveWorkSetting: %+v", errDetail)
	}
	if resp := mustMissing(t, f, admin, "2026-09-30"); resp.From != "2026-09-01" {
		t.Errorf("old start: from = %s, want 2026-09-01", resp.From)
	}

	// a start date after "to" checks nothing yet
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-10-05")}); errDetail != nil {
		t.Fatalf("SaveWorkSetting: %+v", errDetail)
	}
	if resp := mustMissing(t, f, admin, "2026-09-30"); resp.Workdays != 0 || len(resp.Members) != 0 {
		t.Errorf("future start = %+v", resp)
	}
}

func TestSaveWorkSettingUpdatesOnlyGivenFields(t *testing.T) {
	f := newWorkFixture(t)

	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{AllowViewAll: boolPtr(true)}); errDetail != nil {
		t.Fatal(errDetail)
	}
	resp, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("2026-09-01")})
	if errDetail != nil || !resp.AllowViewAll || resp.StartDate != "2026-09-01" {
		t.Fatalf("SaveWorkSetting = %+v, %+v", resp, errDetail)
	}

	options, _ := f.p.GetWorkOptions()
	if options.StartDate != "2026-09-01" || !options.AllowViewAll {
		t.Errorf("options = %+v", options)
	}

	// an empty start date clears it
	if resp, _ := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("")}); resp.StartDate != "" {
		t.Errorf("start date not cleared: %+v", resp)
	}

	_, errDetail = f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{StartDate: strPtr("09/01")})
	expectStatus(t, errDetail, http.StatusBadRequest)
}
