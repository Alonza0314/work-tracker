package processor

import (
	"backend/constant"
	"backend/internal/context"
	"backend/model"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	holidaySyncInterval = 24 * time.Hour

	// names for office calendar entries without a description
	defaultHolidayName = "放假"
	defaultWorkdayName = "補班"
)

var errCalendarNotPublished = errors.New("office calendar not published")

// week summary

// GetWeekSummary sums the account's records for the 7 days from req.From and
// the hours it should log: WORK_DAILY_HOURS per workday, where workdays are
// Monday-Friday minus holidays plus makeup workdays.
func (p *Processor) GetWeekSummary(acc *model.Account, req *model.RequestWeekSummary) (*model.ResponseWeekSummary, *model.ErrorDetail) {
	if !isWorkDate(req.From) {
		return nil, errBadRequest("From must be a YYYY-MM-DD date")
	}
	start, _ := time.Parse(constant.WORK_DATE_LAYOUT, req.From)
	to := start.AddDate(0, 0, 6).Format(constant.WORK_DATE_LAYOUT)

	holidays, errDetail := p.effectiveHolidays(req.From, to)
	if errDetail != nil {
		return nil, errDetail
	}
	records, err := p.ListWorkRecords(&model.WorkRecordFilter{
		Account: acc.Account,
		From:    req.From,
		To:      to,
	})
	if err != nil {
		p.ProcLog.Errorf("Failed to list work records of %s: %v", acc.Account, err)
		return nil, errInternal("Failed to list work records")
	}

	logged := map[string]float64{}
	for _, record := range records {
		logged[record.Date] += record.Hours
	}

	summary := &model.ResponseWeekSummary{
		Message:     "Get week summary successful",
		From:        req.From,
		To:          to,
		RecordCount: len(records),
		Days:        make([]model.WeekDay, 0, 7),
	}
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		date := day.Format(constant.WORK_DATE_LAYOUT)

		workday, name := workdayOf(day, holidays)
		if isWeekday(day) && !workday {
			summary.DaysOff++
		}

		required := 0.0
		if workday {
			required = constant.WORK_DAILY_HOURS
		}
		summary.Days = append(summary.Days, model.WeekDay{
			Date:          date,
			Workday:       workday,
			HolidayName:   name,
			LoggedHours:   roundHours(logged[date]),
			RequiredHours: required,
		})
		summary.LoggedHours += logged[date]
		summary.RequiredHours += required
	}
	summary.LoggedHours = roundHours(summary.LoggedHours)
	summary.RemainingHours = math.Max(0, roundHours(summary.RequiredHours-summary.LoggedHours))

	return summary, nil
}

// ListMissingEntries lists, for everyone but the system admin, the workdays
// of the WORK_MISSING_DAYS days to req.To without any work record (todos do
// not count). The range starts no earlier than the work start date, and days
// before an account was created are skipped. Allowed like everyone's work
// records (checkViewAll).
func (p *Processor) ListMissingEntries(acc *model.Account, req *model.RequestMissingEntries) (*model.ResponseMissingEntries, *model.ErrorDetail) {
	if errDetail := p.checkViewAll(acc); errDetail != nil {
		return nil, errDetail
	}
	if !isWorkDate(req.To) {
		return nil, errBadRequest("To must be a YYYY-MM-DD date")
	}
	end, _ := time.Parse(constant.WORK_DATE_LAYOUT, req.To)
	start := end.AddDate(0, 0, -(constant.WORK_MISSING_DAYS - 1))

	setting, err := p.GetWorkSetting()
	if err != nil {
		p.ProcLog.Errorf("Failed to get work setting: %v", err)
		return nil, errInternal("Failed to get work setting")
	}
	if setting.StartDate != "" {
		if workStart, err := time.Parse(constant.WORK_DATE_LAYOUT, setting.StartDate); err == nil && workStart.After(start) {
			start = workStart
		}
	}
	from := start.Format(constant.WORK_DATE_LAYOUT)

	holidays, errDetail := p.effectiveHolidays(from, req.To)
	if errDetail != nil {
		return nil, errDetail
	}
	var workdays []string
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if workday, _ := workdayOf(day, holidays); workday {
			workdays = append(workdays, day.Format(constant.WORK_DATE_LAYOUT))
		}
	}

	records, err := p.ListWorkRecords(&model.WorkRecordFilter{From: from, To: req.To})
	if err != nil {
		p.ProcLog.Errorf("Failed to list work records: %v", err)
		return nil, errInternal("Failed to list work records")
	}
	logged := map[string]bool{}
	for _, record := range records {
		logged[record.Account+"|"+record.Date] = true
	}

	accounts, err := p.ListAccounts()
	if err != nil {
		p.ProcLog.Errorf("Failed to list accounts: %v", err)
		return nil, errInternal("Failed to list accounts")
	}

	resp := &model.ResponseMissingEntries{
		Message:  "List missing entries successful",
		From:     from,
		To:       req.To,
		Workdays: len(workdays),
		Members:  []model.MissingMember{},
	}
	for _, member := range accounts {
		if member.IsSystem {
			continue
		}
		resp.CheckedCount++

		joined := ""
		if !member.CreatedAt.IsZero() {
			joined = member.CreatedAt.In(time.Local).Format(constant.WORK_DATE_LAYOUT)
		}
		missing := []string{}
		for _, date := range workdays {
			if date >= joined && !logged[member.Account+"|"+date] {
				missing = append(missing, date)
			}
		}
		if len(missing) > 0 {
			resp.Members = append(resp.Members, model.MissingMember{
				Account:      member.Account,
				Name:         member.Name,
				MissingCount: len(missing),
				MissingDates: missing,
			})
		}
	}
	sort.SliceStable(resp.Members, func(i, j int) bool {
		if resp.Members[i].MissingCount != resp.Members[j].MissingCount {
			return resp.Members[i].MissingCount > resp.Members[j].MissingCount
		}
		return resp.Members[i].Account < resp.Members[j].Account
	})

	return resp, nil
}

func isWeekday(day time.Time) bool {
	return day.Weekday() != time.Saturday && day.Weekday() != time.Sunday
}

// workdayOf reports whether day is a workday (Monday-Friday unless the
// calendar says otherwise) and the name of its calendar entry, if any.
func workdayOf(day time.Time, holidays map[string]*model.Holiday) (bool, string) {
	if holiday, ok := holidays[day.Format(constant.WORK_DATE_LAYOUT)]; ok {
		return holiday.Type == constant.HOLIDAY_TYPE_WORKDAY, holiday.Name
	}
	return isWeekday(day), ""
}

// effectiveHolidays maps each date in from..to to its entry, the manual one
// winning over gov.
func (p *Processor) effectiveHolidays(from, to string) (map[string]*model.Holiday, *model.ErrorDetail) {
	holidays, err := p.ListHolidays(from, to)
	if err != nil {
		p.ProcLog.Errorf("Failed to list holidays: %v", err)
		return nil, errInternal("Failed to list holidays")
	}

	byDate := map[string]*model.Holiday{}
	for _, holiday := range holidays {
		if _, taken := byDate[holiday.Date]; !taken || holiday.Source == constant.HOLIDAY_SOURCE_MANUAL {
			byDate[holiday.Date] = holiday
		}
	}
	return byDate, nil
}

// holiday calendar

func (p *Processor) ListYearHolidays(req *model.RequestListHolidays) (*model.ResponseHolidays, *model.ErrorDetail) {
	if year, err := strconv.Atoi(req.Year); err != nil || len(req.Year) != 4 || year < 1 {
		return nil, errBadRequest("Year must have 4 digits")
	}

	byDate, errDetail := p.effectiveHolidays(req.Year+"-01-01", req.Year+"-12-31")
	if errDetail != nil {
		return nil, errDetail
	}
	holidays := make([]model.Holiday, 0, len(byDate))
	for _, holiday := range byDate {
		holidays = append(holidays, *holiday)
	}
	sort.Slice(holidays, func(i, j int) bool { return holidays[i].Date < holidays[j].Date })

	syncedAt, err := p.GetHolidaySyncedAt()
	if err != nil {
		p.ProcLog.Errorf("Failed to get holiday sync time: %v", err)
		return nil, errInternal("Failed to get holiday sync time")
	}

	return &model.ResponseHolidays{
		Message:      "List holidays successful",
		Holidays:     holidays,
		LastSyncedAt: optionalTime(syncedAt),
	}, nil
}

// SaveManualHoliday creates or replaces the admin's entry on date; it wins
// over the gov entry of that date.
func (p *Processor) SaveManualHoliday(date string, req *model.RequestSaveHoliday) (*model.ResponseHoliday, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing save holiday %s", date)

	if !isWorkDate(date) {
		return nil, errBadRequest("Date must be YYYY-MM-DD")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errBadRequest("Name must not be blank")
	}

	holiday := &model.Holiday{
		Date:   date,
		Name:   name,
		Type:   req.Type,
		Source: constant.HOLIDAY_SOURCE_MANUAL,
	}
	if err := p.SaveHoliday(holiday); err != nil {
		p.ProcLog.Errorf("Failed to save holiday %s: %v", date, err)
		return nil, errInternal("Failed to save holiday")
	}

	return &model.ResponseHoliday{
		Message: "Save holiday successful",
		Holiday: holiday,
	}, nil
}

// DeleteManualHoliday removes the admin's entry on date; a gov entry of that
// date applies again.
func (p *Processor) DeleteManualHoliday(date string) (*model.ResponseDeleteHoliday, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing delete holiday %s", date)

	if err := p.DeleteHoliday(date, constant.HOLIDAY_SOURCE_MANUAL); err != nil {
		if errors.Is(err, context.ErrHolidayNotFound) {
			return nil, errNotFound("Holiday not found")
		}
		p.ProcLog.Errorf("Failed to delete holiday %s: %v", date, err)
		return nil, errInternal("Failed to delete holiday")
	}

	return &model.ResponseDeleteHoliday{
		Message: "Delete holiday successful",
	}, nil
}

// sync from the government office calendar

// SyncHolidaysNow syncs right away (the admin's sync button).
func (p *Processor) SyncHolidaysNow() (*model.ResponseSyncHolidays, *model.ErrorDetail) {
	synced, at, err := p.syncHolidays()
	if err != nil {
		p.ProcLog.Warnf("Failed to sync holidays: %v", err)
		return nil, &model.ErrorDetail{
			HttpStatus: http.StatusBadGateway,
			Detail:     "Failed to sync holidays from the office calendar",
		}
	}

	return &model.ResponseSyncHolidays{
		Message:      "Sync holidays successful",
		Synced:       synced,
		LastSyncedAt: &at,
	}, nil
}

// StartHolidaySync syncs in the background now and every day, when enabled
// in the config. Release stops it.
func (p *Processor) StartHolidaySync() {
	if !p.holidaySync || p.holidaySourceUrl == "" || p.holidaySyncStop != nil {
		return
	}
	p.holidaySyncStop = make(chan struct{})
	p.holidaySyncDone = make(chan struct{})

	go func() {
		defer close(p.holidaySyncDone)

		ticker := time.NewTicker(holidaySyncInterval)
		defer ticker.Stop()
		for {
			if synced, _, err := p.syncHolidays(); err != nil {
				p.ProcLog.Warnf("Failed to sync holidays: %v", err)
			} else {
				p.ProcLog.Infof("Synced %d holidays from the office calendar", synced)
			}

			select {
			case <-p.holidaySyncStop:
				return
			case <-ticker.C:
			}
		}
	}()
}

func (p *Processor) stopHolidaySync() {
	if p.holidaySyncStop == nil {
		return
	}
	close(p.holidaySyncStop)
	<-p.holidaySyncDone
	p.holidaySyncStop = nil
}

// syncHolidays replaces the gov entries of this year and the next with the
// office calendar. A next year that is not published yet is skipped.
func (p *Processor) syncHolidays() (int, time.Time, error) {
	now := p.now()
	synced := 0
	for _, year := range []int{now.Year(), now.Year() + 1} {
		holidays, err := p.fetchOfficeCalendar(year)
		if errors.Is(err, errCalendarNotPublished) && year > now.Year() {
			continue
		}
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("year %d: %v", year, err)
		}

		stored, err := p.ReplaceGovHolidays(strconv.Itoa(year), holidays)
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("failed to store holidays of %d: %v", year, err)
		}
		synced += stored
	}

	if err := p.SetHolidaySyncedAt(now); err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to store holiday sync time: %v", err)
	}
	return synced, now, nil
}

// officeCalendarDay is one day of the TaiwanCalendar JSON (from the DGPA
// office calendar): every date of the year with whether it is a day off.
type officeCalendarDay struct {
	Date        string `json:"date"`
	IsHoliday   bool   `json:"isHoliday"`
	Description string `json:"description"`
}

// fetchOfficeCalendar returns the year's exceptions to the Monday-Friday week:
// weekdays off and working weekend days.
func (p *Processor) fetchOfficeCalendar(year int) ([]*model.Holiday, error) {
	url := strings.ReplaceAll(p.holidaySourceUrl, "{year}", strconv.Itoa(year))
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errCalendarNotPublished
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch %s: status %d", url, resp.StatusCode)
	}

	var days []officeCalendarDay
	if err := json.NewDecoder(resp.Body).Decode(&days); err != nil {
		return nil, fmt.Errorf("failed to decode %s: %v", url, err)
	}

	holidays := []*model.Holiday{}
	for _, day := range days {
		date, err := time.Parse("20060102", day.Date)
		if err != nil {
			return nil, fmt.Errorf("bad date %q in %s", day.Date, url)
		}
		weekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

		var kind, name string
		switch {
		case !weekend && day.IsHoliday:
			kind, name = constant.HOLIDAY_TYPE_HOLIDAY, defaultHolidayName
		case weekend && !day.IsHoliday:
			kind, name = constant.HOLIDAY_TYPE_WORKDAY, defaultWorkdayName
		default:
			continue
		}
		if day.Description != "" {
			name = day.Description
		}

		holidays = append(holidays, &model.Holiday{
			Date:   date.Format(constant.WORK_DATE_LAYOUT),
			Name:   name,
			Type:   kind,
			Source: constant.HOLIDAY_SOURCE_GOV,
		})
	}
	return holidays, nil
}

func roundHours(hours float64) float64 {
	return math.Round(hours*100) / 100
}

func optionalTime(at time.Time) *time.Time {
	if at.IsZero() {
		return nil
	}
	return &at
}
