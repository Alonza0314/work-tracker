package context

import (
	"backend/model"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestBboltDb(t *testing.T) *bboltDb {
	t.Helper()

	db, err := newBboltDb(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("newBboltDb: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Release(); err != nil {
			t.Errorf("Release: %v", err)
		}
	})

	return db
}

func testAccount(account string) *model.Account {
	return &model.Account{
		Account:  account,
		Password: "hash",
		Role:     "default",
		I18n:     "en",
	}
}

func TestBboltCreateAndGetAccount(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	got, err := db.GetAccount("alice")
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if *got != *testAccount("alice") {
		t.Errorf("GetAccount = %+v, want %+v", *got, *testAccount("alice"))
	}
}

func TestBboltCreateAccountRejectsDuplicate(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if err := db.CreateAccount(testAccount("alice")); !errors.Is(err, ErrAccountExists) {
		t.Errorf("second CreateAccount err = %v, want ErrAccountExists", err)
	}
}

func TestBboltGetAccountNotFound(t *testing.T) {
	db := newTestBboltDb(t)

	if _, err := db.GetAccount("nobody"); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("GetAccount err = %v, want ErrAccountNotFound", err)
	}
}

func TestBboltUpdateAccount(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	updated := testAccount("alice")
	updated.Role = "admin"
	if err := db.UpdateAccount(updated); err != nil {
		t.Fatalf("UpdateAccount: %v", err)
	}

	got, err := db.GetAccount("alice")
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q, want admin", got.Role)
	}
}

func TestBboltUpdateAccountNotFound(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.UpdateAccount(testAccount("nobody")); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("UpdateAccount err = %v, want ErrAccountNotFound", err)
	}
}

func TestBboltDeleteAccount(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if err := db.DeleteAccount("alice"); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	if _, err := db.GetAccount("alice"); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("GetAccount after delete err = %v, want ErrAccountNotFound", err)
	}
	if err := db.DeleteAccount("alice"); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("second DeleteAccount err = %v, want ErrAccountNotFound", err)
	}
}

func TestBboltListAccountsSortedByAccount(t *testing.T) {
	db := newTestBboltDb(t)

	for _, account := range []string{"carol", "alice", "bob"} {
		if err := db.CreateAccount(testAccount(account)); err != nil {
			t.Fatalf("CreateAccount(%s): %v", account, err)
		}
	}

	accounts, err := db.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	var got []string
	for _, acc := range accounts {
		got = append(got, acc.Account)
	}
	want := []string{"alice", "bob", "carol"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("ListAccounts = %v, want %v", got, want)
	}
}

func TestBboltAccountsPersistAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	db, err := newBboltDb(path)
	if err != nil {
		t.Fatalf("newBboltDb: %v", err)
	}
	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if err := db.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	reopened, err := newBboltDb(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = reopened.Release() }()

	if _, err := reopened.GetAccount("alice"); err != nil {
		t.Errorf("GetAccount after reopen: %v", err)
	}
}

func TestBboltCategoryCreateAssignsIncreasingIDs(t *testing.T) {
	db := newTestBboltDb(t)

	first := &model.WorkOption{Name: "Meeting", Active: true}
	second := &model.WorkOption{Name: "Develop", Active: true}
	if err := db.CreateCategory(first); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if err := db.CreateCategory(second); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if first.ID == "" || first.ID >= second.ID {
		t.Errorf("IDs = %q, %q, want non-empty and increasing", first.ID, second.ID)
	}

	list, err := db.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(list) != 2 || list[0].Name != "Meeting" || list[1].Name != "Develop" {
		t.Errorf("ListCategories = %+v", list)
	}
}

func TestBboltCategoryUpdateAndNotFound(t *testing.T) {
	db := newTestBboltDb(t)

	c := &model.WorkOption{Name: "Meeting", Active: true}
	if err := db.CreateCategory(c); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	c.Active = false
	if err := db.UpdateCategory(c); err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	got, err := db.GetCategory(c.ID)
	if err != nil || got.Active {
		t.Errorf("GetCategory = %+v, %v", got, err)
	}

	if _, err := db.GetCategory("missing"); !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("GetCategory missing err = %v", err)
	}
	if err := db.UpdateCategory(&model.WorkOption{ID: "missing"}); !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("UpdateCategory missing err = %v", err)
	}
}

func TestBboltProjectsAreSeparateFromCategories(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateProject(&model.WorkOption{Name: "WT", Active: true}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	projects, err := db.ListProjects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects = %+v, %v", projects, err)
	}
	categories, err := db.ListCategories()
	if err != nil || len(categories) != 0 {
		t.Errorf("ListCategories = %+v, %v", categories, err)
	}
	if _, err := db.GetProject("missing"); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("GetProject missing err = %v", err)
	}
}

func testRecord(account, date, category, project string) *model.WorkRecord {
	return &model.WorkRecord{
		Account: account,
		WorkEntry: model.WorkEntry{
			Date:        date,
			CategoryID:  category,
			Description: "work",
			Hours:       1.5,
			ProjectID:   project,
		},
		CreatedAt: time.Now(),
	}
}

func TestBboltWorkRecordCRUD(t *testing.T) {
	db := newTestBboltDb(t)

	r := testRecord("alice", "2026-09-01", "c1", "p1")
	if err := db.CreateWorkRecord(r); err != nil {
		t.Fatalf("CreateWorkRecord: %v", err)
	}
	if r.ID == "" {
		t.Fatal("CreateWorkRecord did not assign an ID")
	}

	r.Hours = 3
	if err := db.UpdateWorkRecord(r); err != nil {
		t.Fatalf("UpdateWorkRecord: %v", err)
	}
	got, err := db.GetWorkRecord(r.ID)
	if err != nil || got.Hours != 3 || got.Account != "alice" {
		t.Errorf("GetWorkRecord = %+v, %v", got, err)
	}

	if err := db.DeleteWorkRecord(r.ID); err != nil {
		t.Fatalf("DeleteWorkRecord: %v", err)
	}
	if _, err := db.GetWorkRecord(r.ID); !errors.Is(err, ErrWorkRecordNotFound) {
		t.Errorf("GetWorkRecord after delete err = %v", err)
	}
	if err := db.DeleteWorkRecord(r.ID); !errors.Is(err, ErrWorkRecordNotFound) {
		t.Errorf("second DeleteWorkRecord err = %v", err)
	}
	if err := db.UpdateWorkRecord(r); !errors.Is(err, ErrWorkRecordNotFound) {
		t.Errorf("UpdateWorkRecord after delete err = %v", err)
	}
}

func TestBboltListWorkRecordsFilter(t *testing.T) {
	db := newTestBboltDb(t)

	for _, r := range []*model.WorkRecord{
		testRecord("alice", "2026-09-01", "c1", "p1"),
		testRecord("alice", "2026-09-05", "c2", "p1"),
		testRecord("bob", "2026-09-03", "c1", "p2"),
		testRecord("bob", "2026-09-10", "c2", "p2"),
	} {
		if err := db.CreateWorkRecord(r); err != nil {
			t.Fatalf("CreateWorkRecord: %v", err)
		}
	}

	cases := []struct {
		name   string
		filter model.WorkRecordFilter
		want   int
	}{
		{"all", model.WorkRecordFilter{}, 4},
		{"account", model.WorkRecordFilter{Account: "alice"}, 2},
		{"date range inclusive", model.WorkRecordFilter{From: "2026-09-03", To: "2026-09-05"}, 2},
		{"from only", model.WorkRecordFilter{From: "2026-09-05"}, 2},
		{"category", model.WorkRecordFilter{CategoryID: "c1"}, 2},
		{"project and account", model.WorkRecordFilter{Account: "bob", ProjectID: "p2"}, 2},
		{"no match", model.WorkRecordFilter{Account: "alice", ProjectID: "p2"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.ListWorkRecords(&tc.filter)
			if err != nil {
				t.Fatalf("ListWorkRecords: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("len = %d, want %d (%+v)", len(got), tc.want, got)
			}
		})
	}
}

func testTodo(account string) *model.Todo {
	return &model.Todo{
		Account: account,
		WorkEntry: model.WorkEntry{
			Date:        "2026-09-01",
			CategoryID:  "c1",
			Description: "todo",
			Hours:       2,
			ProjectID:   "p1",
		},
		CreatedAt: time.Now(),
	}
}

func TestBboltTodoCRUDAndListByAccount(t *testing.T) {
	db := newTestBboltDb(t)

	alice := testTodo("alice")
	if err := db.CreateTodo(alice); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if err := db.CreateTodo(testTodo("bob")); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	list, err := db.ListTodos("alice")
	if err != nil || len(list) != 1 || list[0].ID != alice.ID {
		t.Fatalf("ListTodos = %+v, %v", list, err)
	}

	alice.Description = "changed"
	if err := db.UpdateTodo(alice); err != nil {
		t.Fatalf("UpdateTodo: %v", err)
	}
	got, err := db.GetTodo(alice.ID)
	if err != nil || got.Description != "changed" {
		t.Errorf("GetTodo = %+v, %v", got, err)
	}

	if err := db.DeleteTodo(alice.ID); err != nil {
		t.Fatalf("DeleteTodo: %v", err)
	}
	if _, err := db.GetTodo(alice.ID); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("GetTodo after delete err = %v", err)
	}
}

func TestBboltCompleteTodoMovesItToWorkRecords(t *testing.T) {
	db := newTestBboltDb(t)

	todo := testTodo("alice")
	if err := db.CreateTodo(todo); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	record := testRecord("alice", "2026-09-29", "c1", "p1")
	if err := db.CompleteTodo(todo.ID, record); err != nil {
		t.Fatalf("CompleteTodo: %v", err)
	}
	if record.ID == "" {
		t.Error("CompleteTodo did not assign a record ID")
	}
	if _, err := db.GetTodo(todo.ID); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("todo still exists: %v", err)
	}
	if _, err := db.GetWorkRecord(record.ID); err != nil {
		t.Errorf("record not created: %v", err)
	}
}

func TestBboltCompleteMissingTodoCreatesNothing(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CompleteTodo("missing", testRecord("alice", "2026-09-29", "c1", "p1")); !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("CompleteTodo err = %v, want ErrTodoNotFound", err)
	}
	records, err := db.ListWorkRecords(&model.WorkRecordFilter{})
	if err != nil || len(records) != 0 {
		t.Errorf("records = %+v, %v", records, err)
	}
}

func TestBboltWorkSettingDefaultsAndUpdate(t *testing.T) {
	db := newTestBboltDb(t)

	setting, err := db.GetWorkSetting()
	if err != nil || setting.AllowViewAll {
		t.Fatalf("default GetWorkSetting = %+v, %v", setting, err)
	}

	if err := db.UpdateWorkSetting(&model.WorkSetting{AllowViewAll: true}); err != nil {
		t.Fatalf("UpdateWorkSetting: %v", err)
	}
	setting, err = db.GetWorkSetting()
	if err != nil || !setting.AllowViewAll {
		t.Errorf("GetWorkSetting = %+v, %v", setting, err)
	}
}

func TestBboltDeleteCategoryClearsReferences(t *testing.T) {
	db := newTestBboltDb(t)

	keep := &model.WorkOption{Name: "Keep", Active: true}
	drop := &model.WorkOption{Name: "Drop", Active: true}
	for _, c := range []*model.WorkOption{keep, drop} {
		if err := db.CreateCategory(c); err != nil {
			t.Fatalf("CreateCategory: %v", err)
		}
	}

	dropped := testRecord("alice", "2026-09-01", drop.ID, "p1")
	kept := testRecord("alice", "2026-09-02", keep.ID, "p1")
	for _, r := range []*model.WorkRecord{dropped, kept} {
		if err := db.CreateWorkRecord(r); err != nil {
			t.Fatalf("CreateWorkRecord: %v", err)
		}
	}
	todo := testTodo("alice")
	todo.CategoryID = drop.ID
	if err := db.CreateTodo(todo); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	cleared, err := db.DeleteCategory(drop.ID)
	if err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	if cleared != 2 {
		t.Errorf("cleared = %d, want 2", cleared)
	}

	if _, err := db.GetCategory(drop.ID); !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("category still exists: %v", err)
	}
	if got, _ := db.GetWorkRecord(dropped.ID); got.CategoryID != "" || got.ProjectID != "p1" {
		t.Errorf("dropped record = %+v", got)
	}
	if got, _ := db.GetWorkRecord(kept.ID); got.CategoryID != keep.ID {
		t.Errorf("kept record = %+v", got)
	}
	if got, _ := db.GetTodo(todo.ID); got.CategoryID != "" {
		t.Errorf("todo = %+v", got)
	}
}

func TestBboltDeleteProjectClearsReferences(t *testing.T) {
	db := newTestBboltDb(t)

	project := &model.WorkOption{Name: "WT", Active: true}
	if err := db.CreateProject(project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	record := testRecord("alice", "2026-09-01", "c1", project.ID)
	if err := db.CreateWorkRecord(record); err != nil {
		t.Fatalf("CreateWorkRecord: %v", err)
	}

	cleared, err := db.DeleteProject(project.ID)
	if err != nil || cleared != 1 {
		t.Fatalf("DeleteProject = %d, %v", cleared, err)
	}
	if got, _ := db.GetWorkRecord(record.ID); got.ProjectID != "" || got.CategoryID != "c1" {
		t.Errorf("record = %+v", got)
	}
}

func TestBboltDeleteMissingOption(t *testing.T) {
	db := newTestBboltDb(t)

	if _, err := db.DeleteCategory("missing"); !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("DeleteCategory err = %v", err)
	}
	if _, err := db.DeleteProject("missing"); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("DeleteProject err = %v", err)
	}
}

func TestBboltRenameAccountMovesAccountAndEntries(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.CreateAccount(testAccount("alice")); err != nil {
		t.Fatal(err)
	}
	record := testRecord("alice", "2026-09-01", "c1", "p1")
	other := testRecord("bob", "2026-09-01", "c1", "p1")
	for _, r := range []*model.WorkRecord{record, other} {
		if err := db.CreateWorkRecord(r); err != nil {
			t.Fatal(err)
		}
	}
	todo := testTodo("alice")
	if err := db.CreateTodo(todo); err != nil {
		t.Fatal(err)
	}

	if err := db.RenameAccount("alice", "ALICE"); err != nil {
		t.Fatalf("RenameAccount: %v", err)
	}

	if _, err := db.GetAccount("alice"); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("old key still exists: %v", err)
	}
	if acc, err := db.GetAccount("ALICE"); err != nil || acc.Account != "ALICE" || acc.Password != "hash" {
		t.Errorf("new account = %+v, %v", acc, err)
	}
	if got, _ := db.GetWorkRecord(record.ID); got.Account != "ALICE" {
		t.Errorf("record account = %q", got.Account)
	}
	if got, _ := db.GetWorkRecord(other.ID); got.Account != "bob" {
		t.Errorf("other record account = %q", got.Account)
	}
	if got, _ := db.GetTodo(todo.ID); got.Account != "ALICE" {
		t.Errorf("todo account = %q", got.Account)
	}
}

func TestBboltRenameAccountErrors(t *testing.T) {
	db := newTestBboltDb(t)

	if err := db.RenameAccount("nobody", "NOBODY"); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("missing err = %v", err)
	}

	for _, account := range []string{"alice", "ALICE"} {
		if err := db.CreateAccount(testAccount(account)); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.RenameAccount("alice", "ALICE"); !errors.Is(err, ErrAccountExists) {
		t.Errorf("collision err = %v", err)
	}
	if _, err := db.GetAccount("alice"); err != nil {
		t.Errorf("failed rename changed data: %v", err)
	}
}

func testHoliday(date, name, kind, source string) *model.Holiday {
	return &model.Holiday{Date: date, Name: name, Type: kind, Source: source}
}

func TestBboltHolidaysKeepBothSourcesPerDate(t *testing.T) {
	db := newTestBboltDb(t)

	for _, h := range []*model.Holiday{
		testHoliday("2026-10-09", "National Day off", "holiday", "gov"),
		testHoliday("2026-10-09", "Company workday", "workday", "manual"),
		testHoliday("2026-10-10", "National Day", "holiday", "gov"),
		testHoliday("2026-11-01", "Later", "holiday", "gov"),
	} {
		if err := db.SaveHoliday(h); err != nil {
			t.Fatalf("SaveHoliday: %v", err)
		}
	}

	got, err := db.ListHolidays("2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("ListHolidays: %v", err)
	}
	if len(got) != 3 || got[0].Date != "2026-10-09" || got[2].Date != "2026-10-10" {
		t.Errorf("ListHolidays = %+v", got)
	}

	if err := db.DeleteHoliday("2026-10-09", "manual"); err != nil {
		t.Fatalf("DeleteHoliday: %v", err)
	}
	if err := db.DeleteHoliday("2026-10-09", "manual"); !errors.Is(err, ErrHolidayNotFound) {
		t.Errorf("second DeleteHoliday err = %v", err)
	}
	got, _ = db.ListHolidays("2026-10-09", "2026-10-09")
	if len(got) != 1 || got[0].Source != "gov" {
		t.Errorf("after delete = %+v", got)
	}
}

func TestBboltReplaceGovHolidaysOnlyTouchesThatYearsGovEntries(t *testing.T) {
	db := newTestBboltDb(t)

	for _, h := range []*model.Holiday{
		testHoliday("2026-01-01", "Old gov", "holiday", "gov"),
		testHoliday("2026-05-01", "Manual", "holiday", "manual"),
		testHoliday("2027-01-01", "Next year gov", "holiday", "gov"),
	} {
		if err := db.SaveHoliday(h); err != nil {
			t.Fatal(err)
		}
	}

	stored, err := db.ReplaceGovHolidays("2026", []*model.Holiday{
		testHoliday("2026-02-16", "New gov", "holiday", "gov"),
		testHoliday("2026-02-21", "Makeup", "workday", "gov"),
	})
	if err != nil || stored != 2 {
		t.Fatalf("ReplaceGovHolidays = %d, %v", stored, err)
	}

	got, _ := db.ListHolidays("2026-01-01", "2027-12-31")
	var names []string
	for _, h := range got {
		names = append(names, h.Name)
	}
	want := "New gov,Makeup,Manual,Next year gov"
	if strings.Join(names, ",") != want {
		t.Errorf("holidays = %v, want %s", names, want)
	}
}

func TestBboltHolidaySyncedAt(t *testing.T) {
	db := newTestBboltDb(t)

	at, err := db.GetHolidaySyncedAt()
	if err != nil || !at.IsZero() {
		t.Fatalf("default = %v, %v", at, err)
	}

	now := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	if err := db.SetHolidaySyncedAt(now); err != nil {
		t.Fatal(err)
	}
	at, err = db.GetHolidaySyncedAt()
	if err != nil || !at.Equal(now) {
		t.Errorf("GetHolidaySyncedAt = %v, %v", at, err)
	}
}
