package context

import (
	"backend/model"
	"errors"
	"path/filepath"
	"testing"
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
