package processor

import (
	"backend/constant"
	"backend/model"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

// workFixture is a processor with one active category and project, plus a
// default user alice.
type workFixture struct {
	p        *Processor
	alice    *model.Account
	category string
	project  string
}

func newWorkFixture(t *testing.T) *workFixture {
	t.Helper()

	p := newTestProcessor(t)
	alice := mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	category, errDetail := p.CreateWorkCategory(&model.RequestCreateWorkOption{Name: "Meeting"})
	if errDetail != nil {
		t.Fatalf("CreateWorkCategory: %+v", errDetail)
	}
	project, errDetail := p.CreateWorkProject(&model.RequestCreateWorkOption{Name: "Work Tracker"})
	if errDetail != nil {
		t.Fatalf("CreateWorkProject: %+v", errDetail)
	}

	return &workFixture{
		p:        p,
		alice:    alice,
		category: category.Option.ID,
		project:  project.Option.ID,
	}
}

func (f *workFixture) entry(date string, hours float64) *model.RequestSaveWorkEntry {
	return &model.RequestSaveWorkEntry{
		Date:        date,
		CategoryID:  f.category,
		Description: "did something",
		Hours:       hours,
		ProjectID:   f.project,
	}
}

func (f *workFixture) mustCreateRecord(t *testing.T, acc *model.Account, date string, hours float64) *model.WorkRecord {
	t.Helper()

	resp, errDetail := f.p.CreateMyWorkRecord(acc, f.entry(date, hours))
	if errDetail != nil {
		t.Fatalf("CreateMyWorkRecord: %+v", errDetail)
	}
	return resp.Record
}

// allMyRecords lists every record of the account (a range wider than any test data)
func allMyRecords() *model.RequestListMyWorkRecords {
	return &model.RequestListMyWorkRecords{From: "2000-01-01", To: "2100-12-31"}
}

func boolPtr(b bool) *bool {
	return &b
}

// options & settings

func TestCreateCategoryRejectsDuplicateName(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.CreateWorkCategory(&model.RequestCreateWorkOption{Name: " meeting "})
	expectStatus(t, errDetail, http.StatusConflict)
}

func TestCreateCategoryRejectsBlankName(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.CreateWorkCategory(&model.RequestCreateWorkOption{Name: "  "})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestUpdateProjectRenamesAndDeactivates(t *testing.T) {
	f := newWorkFixture(t)

	resp, errDetail := f.p.UpdateWorkProject(f.project, &model.RequestUpdateWorkOption{Name: strPtr("WT"), Active: boolPtr(false)})
	if errDetail != nil {
		t.Fatalf("UpdateWorkProject: %+v", errDetail)
	}
	if resp.Option.Name != "WT" || resp.Option.Active {
		t.Errorf("option = %+v", resp.Option)
	}
}

func TestUpdateCategoryNotFound(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.UpdateWorkCategory("missing", &model.RequestUpdateWorkOption{Active: boolPtr(false)})
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestGetWorkOptionsIncludesInactiveAndSetting(t *testing.T) {
	f := newWorkFixture(t)

	if _, errDetail := f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Active: boolPtr(false)}); errDetail != nil {
		t.Fatalf("UpdateWorkCategory: %+v", errDetail)
	}
	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{AllowViewAll: boolPtr(true)}); errDetail != nil {
		t.Fatalf("SaveWorkSetting: %+v", errDetail)
	}

	resp, errDetail := f.p.GetWorkOptions()
	if errDetail != nil {
		t.Fatalf("GetWorkOptions: %+v", errDetail)
	}
	if len(resp.Categories) != 1 || resp.Categories[0].Active || len(resp.Projects) != 1 || !resp.AllowViewAll {
		t.Errorf("options = %+v", resp)
	}
}

// my work records

func TestCreateMyWorkRecordValidatesEntry(t *testing.T) {
	f := newWorkFixture(t)

	cases := map[string]func(*model.RequestSaveWorkEntry){
		"bad date":          func(e *model.RequestSaveWorkEntry) { e.Date = "2026/09/01" },
		"zero hours":        func(e *model.RequestSaveWorkEntry) { e.Hours = 0 },
		"too many hours":    func(e *model.RequestSaveWorkEntry) { e.Hours = 24.5 },
		"not half hours":    func(e *model.RequestSaveWorkEntry) { e.Hours = 1.3 },
		"blank description": func(e *model.RequestSaveWorkEntry) { e.Description = "  " },
		"unknown category":  func(e *model.RequestSaveWorkEntry) { e.CategoryID = "missing" },
		"unknown project":   func(e *model.RequestSaveWorkEntry) { e.ProjectID = "missing" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			entry := f.entry("2026-09-01", 1)
			mutate(entry)
			_, errDetail := f.p.CreateMyWorkRecord(f.alice, entry)
			expectStatus(t, errDetail, http.StatusBadRequest)
		})
	}
}

func TestCreateMyWorkRecordRequiresCategoryAndHours(t *testing.T) {
	f := newWorkFixture(t)

	noCategory := f.entry("2026-09-01", 1)
	noCategory.CategoryID = ""
	_, errDetail := f.p.CreateMyWorkRecord(f.alice, noCategory)
	expectStatus(t, errDetail, http.StatusBadRequest)

	_, errDetail = f.p.CreateMyWorkRecord(f.alice, f.entry("2026-09-01", 0))
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestCreateMyWorkRecordProjectIsOptional(t *testing.T) {
	f := newWorkFixture(t)

	entry := f.entry("2026-09-01", 1)
	entry.ProjectID = ""
	resp, errDetail := f.p.CreateMyWorkRecord(f.alice, entry)
	if errDetail != nil {
		t.Fatalf("CreateMyWorkRecord: %+v", errDetail)
	}
	if resp.Record.ProjectID != "" {
		t.Errorf("projectId = %q, want empty", resp.Record.ProjectID)
	}
}

func TestCreateMyWorkRecordRejectsInactiveCategory(t *testing.T) {
	f := newWorkFixture(t)
	if _, errDetail := f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Active: boolPtr(false)}); errDetail != nil {
		t.Fatalf("UpdateWorkCategory: %+v", errDetail)
	}

	_, errDetail := f.p.CreateMyWorkRecord(f.alice, f.entry("2026-09-01", 1))
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestUpdateMyWorkRecordKeepsInactiveCategoryItAlreadyHas(t *testing.T) {
	f := newWorkFixture(t)
	record := f.mustCreateRecord(t, f.alice, "2026-09-01", 1)
	if _, errDetail := f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Active: boolPtr(false)}); errDetail != nil {
		t.Fatalf("UpdateWorkCategory: %+v", errDetail)
	}

	resp, errDetail := f.p.UpdateMyWorkRecord(f.alice, record.ID, f.entry("2026-09-02", 3))
	if errDetail != nil {
		t.Fatalf("UpdateMyWorkRecord: %+v", errDetail)
	}
	if resp.Record.Date != "2026-09-02" || resp.Record.Hours != 3 || resp.Record.Account != "ALICE" {
		t.Errorf("record = %+v", resp.Record)
	}
}

func TestUpdateAndDeleteOthersWorkRecordNotFound(t *testing.T) {
	f := newWorkFixture(t)
	record := f.mustCreateRecord(t, f.alice, "2026-09-01", 1)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_ADMIN)

	_, errDetail := f.p.UpdateMyWorkRecord(bob, record.ID, f.entry("2026-09-01", 2))
	expectStatus(t, errDetail, http.StatusNotFound)

	_, errDetail = f.p.DeleteMyWorkRecord(bob, record.ID)
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestDeleteMyWorkRecord(t *testing.T) {
	f := newWorkFixture(t)
	record := f.mustCreateRecord(t, f.alice, "2026-09-01", 1)

	if _, errDetail := f.p.DeleteMyWorkRecord(f.alice, record.ID); errDetail != nil {
		t.Fatalf("DeleteMyWorkRecord: %+v", errDetail)
	}
	resp, _ := f.p.ListMyWorkRecords(f.alice, allMyRecords())
	if resp.Total != 0 {
		t.Errorf("total = %d after delete", resp.Total)
	}
}

func TestListMyWorkRecordsInRangeNewestFirst(t *testing.T) {
	f := newWorkFixture(t)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_DEFAULT)
	f.mustCreateRecord(t, bob, "2026-09-30", 1)

	// 25 records over 5 days, 5 per day
	for i := 0; i < 25; i++ {
		f.mustCreateRecord(t, f.alice, fmt.Sprintf("2026-09-%02d", 1+i/5), 1)
	}
	latest := f.mustCreateRecord(t, f.alice, "2026-09-05", 2)
	f.mustCreateRecord(t, f.alice, "2026-09-06", 4)

	week, errDetail := f.p.ListMyWorkRecords(f.alice, &model.RequestListMyWorkRecords{From: "2026-09-01", To: "2026-09-05"})
	if errDetail != nil {
		t.Fatalf("ListMyWorkRecords: %+v", errDetail)
	}
	// no page size limit: every record in the range
	if week.Total != 26 || len(week.Records) != 26 || week.TotalHours != 27 {
		t.Fatalf("range = total %d len %d hours %v", week.Total, len(week.Records), week.TotalHours)
	}
	if week.Records[0].ID != latest.ID {
		t.Errorf("first record = %+v, want the latest created on the newest date", week.Records[0])
	}
	for i := 1; i < len(week.Records); i++ {
		if week.Records[i].Date > week.Records[i-1].Date {
			t.Fatalf("records not sorted by date desc at %d", i)
		}
	}

	day, _ := f.p.ListMyWorkRecords(f.alice, &model.RequestListMyWorkRecords{From: "2026-09-06", To: "2026-09-06"})
	if day.Total != 1 || day.TotalHours != 4 {
		t.Errorf("single day = %+v", day)
	}
}

func TestListWorkRecordsRequireValidRange(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)

	for name, req := range map[string]*model.RequestListMyWorkRecords{
		"missing to":   {From: "2026-09-01"},
		"missing from": {To: "2026-09-01"},
		"bad date":     {From: "09/01", To: "2026-09-07"},
		"reversed":     {From: "2026-09-07", To: "2026-09-01"},
	} {
		_, errDetail := f.p.ListMyWorkRecords(f.alice, req)
		if errDetail == nil || errDetail.HttpStatus != http.StatusBadRequest {
			t.Errorf("my %s: %+v", name, errDetail)
		}
		_, errDetail = f.p.ListAllWorkRecords(admin, &model.RequestListWorkRecords{From: req.From, To: req.To})
		if errDetail == nil || errDetail.HttpStatus != http.StatusBadRequest {
			t.Errorf("all %s: %+v", name, errDetail)
		}
	}
}

// todos

func TestTodoLifecycle(t *testing.T) {
	f := newWorkFixture(t)

	created, errDetail := f.p.CreateMyTodo(f.alice, f.entry("2026-10-02", 2))
	if errDetail != nil {
		t.Fatalf("CreateMyTodo: %+v", errDetail)
	}
	if _, errDetail := f.p.CreateMyTodo(f.alice, f.entry("2026-10-01", 1)); errDetail != nil {
		t.Fatalf("CreateMyTodo: %+v", errDetail)
	}

	list, _ := f.p.ListMyTodos(f.alice)
	if len(list.Todos) != 2 || list.Todos[0].Date != "2026-10-01" {
		t.Fatalf("todos not sorted by date: %+v", list.Todos)
	}

	updated, errDetail := f.p.UpdateMyTodo(f.alice, created.Todo.ID, f.entry("2026-10-03", 4))
	if errDetail != nil || updated.Todo.Hours != 4 {
		t.Fatalf("UpdateMyTodo = %+v, %+v", updated, errDetail)
	}

	if _, errDetail := f.p.DeleteMyTodo(f.alice, created.Todo.ID); errDetail != nil {
		t.Fatalf("DeleteMyTodo: %+v", errDetail)
	}
	list, _ = f.p.ListMyTodos(f.alice)
	if len(list.Todos) != 1 {
		t.Errorf("todos after delete = %+v", list.Todos)
	}
}

func TestCompleteMyTodoCreatesRecordOnGivenDate(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.p.CreateMyTodo(f.alice, f.entry("2026-10-01", 2))

	resp, errDetail := f.p.CompleteMyTodo(f.alice, todo.Todo.ID, &model.RequestCompleteTodo{Date: "2026-09-29"})
	if errDetail != nil {
		t.Fatalf("CompleteMyTodo: %+v", errDetail)
	}
	if resp.Record.Date != "2026-09-29" || resp.Record.Hours != 2 || resp.Record.Account != "ALICE" || resp.Record.Description != "did something" {
		t.Errorf("record = %+v", resp.Record)
	}

	todos, _ := f.p.ListMyTodos(f.alice)
	records, _ := f.p.ListMyWorkRecords(f.alice, allMyRecords())
	if len(todos.Todos) != 0 || records.Total != 1 {
		t.Errorf("todos = %d, records = %d", len(todos.Todos), records.Total)
	}
}

func TestCreateMyTodoOnlyNeedsDateAndDescription(t *testing.T) {
	f := newWorkFixture(t)

	resp, errDetail := f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: "write report"})
	if errDetail != nil {
		t.Fatalf("CreateMyTodo: %+v", errDetail)
	}
	if resp.Todo.CategoryID != "" || resp.Todo.ProjectID != "" || resp.Todo.Hours != 0 {
		t.Errorf("todo = %+v", resp.Todo)
	}

	_, errDetail = f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: " "})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestCreateMyTodoStillValidatesGivenOptions(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: "x", CategoryID: "missing"})
	expectStatus(t, errDetail, http.StatusBadRequest)
	_, errDetail = f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: "x", Hours: 25})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestCompleteMyTodoNeedsCategoryAndHours(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: "write report"})

	_, errDetail := f.p.CompleteMyTodo(f.alice, todo.Todo.ID, &model.RequestCompleteTodo{Date: "2026-09-29"})
	expectStatus(t, errDetail, http.StatusBadRequest)

	todos, _ := f.p.ListMyTodos(f.alice)
	if len(todos.Todos) != 1 {
		t.Fatalf("todo removed after a failed completion")
	}
}

func TestCompleteMyTodoFillsMissingFields(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.p.CreateMyTodo(f.alice, &model.RequestSaveWorkEntry{Date: "2026-10-01", Description: "write report"})

	resp, errDetail := f.p.CompleteMyTodo(f.alice, todo.Todo.ID, &model.RequestCompleteTodo{
		Date:       "2026-09-29",
		CategoryID: f.category,
		Hours:      2.5,
	})
	if errDetail != nil {
		t.Fatalf("CompleteMyTodo: %+v", errDetail)
	}
	if resp.Record.CategoryID != f.category || resp.Record.Hours != 2.5 || resp.Record.ProjectID != "" || resp.Record.Description != "write report" {
		t.Errorf("record = %+v", resp.Record)
	}
}

func TestCompleteMyTodoRejectsBadDate(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.p.CreateMyTodo(f.alice, f.entry("2026-10-01", 2))

	_, errDetail := f.p.CompleteMyTodo(f.alice, todo.Todo.ID, &model.RequestCompleteTodo{Date: "today"})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestOthersTodoNotFound(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.p.CreateMyTodo(f.alice, f.entry("2026-10-01", 2))
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_DEFAULT)

	_, errDetail := f.p.CompleteMyTodo(bob, todo.Todo.ID, &model.RequestCompleteTodo{Date: "2026-09-29"})
	expectStatus(t, errDetail, http.StatusNotFound)
	_, errDetail = f.p.UpdateMyTodo(bob, todo.Todo.ID, f.entry("2026-10-01", 1))
	expectStatus(t, errDetail, http.StatusNotFound)
	_, errDetail = f.p.DeleteMyTodo(bob, todo.Todo.ID)
	expectStatus(t, errDetail, http.StatusNotFound)
}

// everyone's records

func TestListWorkRecordsForbiddenForDefaultUserUnlessAllowed(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.ListAllWorkRecords(f.alice, &model.RequestListWorkRecords{})
	expectStatus(t, errDetail, http.StatusForbidden)
	_, errDetail = f.p.ListWorkMembers(f.alice)
	expectStatus(t, errDetail, http.StatusForbidden)

	if _, errDetail := f.p.SaveWorkSetting(&model.RequestUpdateWorkSetting{AllowViewAll: boolPtr(true)}); errDetail != nil {
		t.Fatalf("SaveWorkSetting: %+v", errDetail)
	}
	if _, errDetail := f.p.ListAllWorkRecords(f.alice, &model.RequestListWorkRecords{From: "2026-09-01", To: "2026-09-07"}); errDetail != nil {
		t.Errorf("ListAllWorkRecords after allow: %+v", errDetail)
	}
}

func TestListWorkRecordsFiltersAndSumsHours(t *testing.T) {
	f := newWorkFixture(t)
	admin := mustGet(t, f.p, testAdminAccount)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_DEFAULT)
	f.mustCreateRecord(t, f.alice, "2026-09-01", 1.5)
	f.mustCreateRecord(t, f.alice, "2026-09-10", 2)
	f.mustCreateRecord(t, bob, "2026-09-05", 4)

	all, errDetail := f.p.ListAllWorkRecords(admin, &model.RequestListWorkRecords{From: "2026-09-01", To: "2026-09-30"})
	if errDetail != nil {
		t.Fatalf("ListAllWorkRecords: %+v", errDetail)
	}
	if all.Total != 3 || all.TotalHours != 7.5 || all.Records[0].Date != "2026-09-10" {
		t.Errorf("all = total %d hours %v first %+v", all.Total, all.TotalHours, all.Records[0])
	}

	filtered, _ := f.p.ListAllWorkRecords(admin, &model.RequestListWorkRecords{Account: "alice", From: "2026-09-02", To: "2026-09-30"})
	if filtered.Total != 1 || filtered.TotalHours != 2 {
		t.Errorf("filtered = total %d hours %v", filtered.Total, filtered.TotalHours)
	}

}

func TestListWorkMembers(t *testing.T) {
	f := newWorkFixture(t)

	resp, errDetail := f.p.ListWorkMembers(mustGet(t, f.p, testAdminAccount))
	if errDetail != nil {
		t.Fatalf("ListWorkMembers: %+v", errDetail)
	}
	if len(resp.Members) != 2 || resp.Members[1].Account != "ALICE" || resp.Members[1].Name != "alice name" {
		t.Errorf("members = %+v", resp.Members)
	}
}

func TestDeleteWorkCategoryClearsEntries(t *testing.T) {
	f := newWorkFixture(t)
	record := f.mustCreateRecord(t, f.alice, "2026-09-01", 1)

	resp, errDetail := f.p.DeleteWorkCategory(f.category)
	if errDetail != nil {
		t.Fatalf("DeleteWorkCategory: %+v", errDetail)
	}
	if resp.Cleared != 1 {
		t.Errorf("cleared = %d, want 1", resp.Cleared)
	}

	options, _ := f.p.GetWorkOptions()
	if len(options.Categories) != 0 {
		t.Errorf("categories = %+v", options.Categories)
	}
	records, _ := f.p.ListMyWorkRecords(f.alice, allMyRecords())
	if records.Records[0].ID != record.ID || records.Records[0].CategoryID != "" {
		t.Errorf("record = %+v", records.Records[0])
	}

	// a record left without a category must get one when edited
	entry := f.entry("2026-09-01", 1)
	entry.CategoryID = ""
	_, errDetail = f.p.UpdateMyWorkRecord(f.alice, record.ID, entry)
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestDeleteWorkProject(t *testing.T) {
	f := newWorkFixture(t)
	f.mustCreateRecord(t, f.alice, "2026-09-01", 1)

	resp, errDetail := f.p.DeleteWorkProject(f.project)
	if errDetail != nil || resp.Cleared != 1 {
		t.Fatalf("DeleteWorkProject = %+v, %+v", resp, errDetail)
	}
}

func TestDeleteWorkOptionNotFound(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.DeleteWorkCategory("missing")
	expectStatus(t, errDetail, http.StatusNotFound)
	_, errDetail = f.p.DeleteWorkProject("missing")
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestWorkHoursAcceptHalfHourSteps(t *testing.T) {
	f := newWorkFixture(t)

	for _, hours := range []float64{0.5, 1.5, 7.5, 24} {
		if _, errDetail := f.p.CreateMyWorkRecord(f.alice, f.entry("2026-09-01", hours)); errDetail != nil {
			t.Errorf("hours %v rejected: %+v", hours, errDetail)
		}
	}

	todo := &model.RequestSaveWorkEntry{Date: "2026-09-01", Description: "x", Hours: 0.25}
	_, errDetail := f.p.CreateMyTodo(f.alice, todo)
	expectStatus(t, errDetail, http.StatusBadRequest)
}

// category colors

func TestCreateCategoryAssignsLeastUsedColor(t *testing.T) {
	f := newWorkFixture(t) // already has one category with the first color

	options, _ := f.p.GetWorkOptions()
	if options.Categories[0].Color != constant.WORK_CATEGORY_COLORS[0] {
		t.Fatalf("first color = %q", options.Categories[0].Color)
	}

	second, errDetail := f.p.CreateWorkCategory(&model.RequestCreateWorkOption{Name: "Develop"})
	if errDetail != nil {
		t.Fatalf("CreateWorkCategory: %+v", errDetail)
	}
	if second.Option.Color != constant.WORK_CATEGORY_COLORS[1] {
		t.Errorf("second color = %q, want %q", second.Option.Color, constant.WORK_CATEGORY_COLORS[1])
	}

	// freeing the first color makes it the least used again
	if _, errDetail := f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Color: strPtr(constant.WORK_CATEGORY_COLORS[1])}); errDetail != nil {
		t.Fatalf("UpdateWorkCategory: %+v", errDetail)
	}
	third, _ := f.p.CreateWorkCategory(&model.RequestCreateWorkOption{Name: "Docs"})
	if third.Option.Color != constant.WORK_CATEGORY_COLORS[0] {
		t.Errorf("third color = %q, want %q", third.Option.Color, constant.WORK_CATEGORY_COLORS[0])
	}
}

func TestUpdateCategoryColorValidatesPalette(t *testing.T) {
	f := newWorkFixture(t)

	resp, errDetail := f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Color: strPtr("purple")})
	if errDetail != nil || resp.Option.Color != "purple" {
		t.Fatalf("UpdateWorkCategory = %+v, %+v", resp, errDetail)
	}

	_, errDetail = f.p.UpdateWorkCategory(f.category, &model.RequestUpdateWorkOption{Color: strPtr("#ff0000")})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestProjectsHaveNoColor(t *testing.T) {
	f := newWorkFixture(t)

	options, _ := f.p.GetWorkOptions()
	if options.Projects[0].Color != "" {
		t.Errorf("project color = %q", options.Projects[0].Color)
	}
	_, errDetail := f.p.UpdateWorkProject(f.project, &model.RequestUpdateWorkOption{Color: strPtr("purple")})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestGetWorkOptionsFillsMissingCategoryColor(t *testing.T) {
	f := newWorkFixture(t)

	// a category stored before categories had colors
	legacy := &model.WorkOption{Name: "Legacy", Active: true}
	if err := f.p.CreateCategory(legacy); err != nil {
		t.Fatal(err)
	}

	first, _ := f.p.GetWorkOptions()
	second, _ := f.p.GetWorkOptions()
	color := first.Categories[1].Color
	if !slices.Contains(constant.WORK_CATEGORY_COLORS, color) || second.Categories[1].Color != color {
		t.Errorf("legacy colors = %q, %q", color, second.Categories[1].Color)
	}
}
