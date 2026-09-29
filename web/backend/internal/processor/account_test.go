package processor

import (
	"backend/constant"
	"backend/internal/context"
	"backend/logger"
	"backend/model"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	loggergoUtil "github.com/Alonza0314/logger-go/v2/util"
	"github.com/free-ran-ue/util"
)

const (
	testAdminAccount  = "admin"
	testAdminPassword = "0000"
	testJwtSecret     = "test-secret"
)

func newTestProcessor(t *testing.T) *Processor {
	t.Helper()
	return newTestProcessorAt(t, filepath.Join(t.TempDir(), "test.db"), testAdminAccount, testAdminPassword)
}

func newTestProcessorAt(t *testing.T, dbPath, adminAccount, adminPassword string) *Processor {
	t.Helper()

	backendLogger := logger.NewBackendLogger(loggergoUtil.LogLevelString("error"), "", true)

	sysCtx := context.NewSystemContext(&context.SystemContextIE{
		DbType: "bbolt",
		DbPath: dbPath,

		BackendLogger: backendLogger,
	})
	if sysCtx == nil {
		t.Fatal("NewSystemContext returned nil")
	}

	p := NewProcessor(&ProcessorIE{
		Username: adminAccount,
		Password: adminPassword,

		JwtSecret:    testJwtSecret,
		JwtExpiresIn: time.Hour,

		SystemContext: sysCtx,

		BackendLogger: backendLogger,
	})
	t.Cleanup(p.Release)

	if err := p.InitSystemAdmin(); err != nil {
		t.Fatalf("InitSystemAdmin: %v", err)
	}

	return p
}

func mustCreateUser(t *testing.T, p *Processor, account, role string) *model.Account {
	t.Helper()

	if _, errDetail := p.CreateUser(&model.RequestCreateUser{
		Account: account,
		Name:    account + " name",
		Role:    role,
		I18n:    constant.I18N_EN,
	}); errDetail != nil {
		t.Fatalf("CreateUser(%s): %+v", account, errDetail)
	}

	acc, err := p.GetAccount(account)
	if err != nil {
		t.Fatalf("GetAccount(%s): %v", account, err)
	}
	return acc
}

func expectStatus(t *testing.T, errDetail *model.ErrorDetail, want int) {
	t.Helper()

	if errDetail == nil {
		t.Fatalf("expected error with status %d, got nil", want)
	}
	if errDetail.HttpStatus != want {
		t.Fatalf("status = %d (%s), want %d", errDetail.HttpStatus, errDetail.Detail, want)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestInitSystemAdminCreatesSystemAccount(t *testing.T) {
	p := newTestProcessor(t)

	acc, err := p.GetAccount(testAdminAccount)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if acc.Role != constant.ROLE_ADMIN || !acc.IsSystem || acc.I18n != constant.DEFAULT_I18N || acc.Name != testAdminAccount {
		t.Errorf("system admin = %+v", acc)
	}
	if acc.Password == testAdminPassword {
		t.Error("password stored in plain text")
	}
}

func TestInitSystemAdminKeepsI18nAndAppliesNewPassword(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	first := newTestProcessorAt(t, dbPath, testAdminAccount, testAdminPassword)
	if _, errDetail := first.UpdateMe(mustGet(t, first, testAdminAccount), &model.RequestUpdateMe{I18n: constant.I18N_EN}); errDetail != nil {
		t.Fatalf("UpdateMe: %+v", errDetail)
	}
	first.Release()

	second := newTestProcessorAt(t, dbPath, testAdminAccount, "new-pw")

	acc := mustGet(t, second, testAdminAccount)
	if acc.I18n != constant.I18N_EN {
		t.Errorf("I18n = %q, want %q", acc.I18n, constant.I18N_EN)
	}
	if _, errDetail := second.Login(&model.RequestLogin{Account: testAdminAccount, Password: "new-pw"}); errDetail != nil {
		t.Errorf("login with new config password: %+v", errDetail)
	}
	if _, errDetail := second.Login(&model.RequestLogin{Account: testAdminAccount, Password: testAdminPassword}); errDetail == nil {
		t.Error("login with old config password succeeded")
	}
}

func TestInitSystemAdminClearsPreviousSystemFlag(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	first := newTestProcessorAt(t, dbPath, "old-admin", testAdminPassword)
	first.Release()

	second := newTestProcessorAt(t, dbPath, testAdminAccount, testAdminPassword)

	old := mustGet(t, second, "old-admin")
	if old.IsSystem {
		t.Error("previous config admin still marked as system")
	}
	if !mustGet(t, second, testAdminAccount).IsSystem {
		t.Error("new config admin not marked as system")
	}
}

func mustGet(t *testing.T, p *Processor, account string) *model.Account {
	t.Helper()

	acc, err := p.GetAccount(account)
	if err != nil {
		t.Fatalf("GetAccount(%s): %v", account, err)
	}
	return acc
}

func TestLoginTokenCarriesRoleAndI18n(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	resp, errDetail := p.Login(&model.RequestLogin{Account: "alice", Password: "alice"})
	if errDetail != nil {
		t.Fatalf("Login: %+v", errDetail)
	}

	claims, err := util.ValidateJWT(resp.Token, testJwtSecret)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if claims["sub"] != "alice" || claims[constant.JWT_CLAIM_NAME] != "alice name" || claims[constant.JWT_CLAIM_ROLE] != constant.ROLE_DEFAULT || claims[constant.JWT_CLAIM_I18N] != constant.I18N_EN {
		t.Errorf("claims = %v", claims)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.Login(&model.RequestLogin{Account: testAdminAccount, Password: "wrong"})
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestLoginRejectsUnknownAccount(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.Login(&model.RequestLogin{Account: "nobody", Password: "x"})
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestCreateUserRejectsDuplicate(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	_, errDetail := p.CreateUser(&model.RequestCreateUser{Account: "alice", Name: "Alice", Role: constant.ROLE_DEFAULT, I18n: constant.I18N_EN})
	expectStatus(t, errDetail, http.StatusConflict)
}

func TestCreateUserRejectsBlankAccount(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.CreateUser(&model.RequestCreateUser{Account: "  ", Name: "Alice", Role: constant.ROLE_DEFAULT, I18n: constant.I18N_EN})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestCreateUserHashesPasswordAndHidesIt(t *testing.T) {
	p := newTestProcessor(t)

	resp, errDetail := p.CreateUser(&model.RequestCreateUser{Account: "alice", Name: "Alice", Role: constant.ROLE_ADMIN, I18n: constant.I18N_EN})
	if errDetail != nil {
		t.Fatalf("CreateUser: %+v", errDetail)
	}
	if resp.User == nil || resp.User.Account != "alice" || resp.User.Name != "Alice" || resp.User.Role != constant.ROLE_ADMIN || resp.User.IsSystem {
		t.Errorf("response user = %+v", resp.User)
	}
	if mustGet(t, p, "alice").Password == "alice" {
		t.Error("password stored in plain text")
	}
}

func TestCreateUserDefaultPasswordIsAccount(t *testing.T) {
	p := newTestProcessor(t)

	if _, errDetail := p.CreateUser(&model.RequestCreateUser{Account: "alice", Name: "Alice", Role: constant.ROLE_DEFAULT, I18n: constant.I18N_EN}); errDetail != nil {
		t.Fatalf("CreateUser: %+v", errDetail)
	}
	if _, errDetail := p.Login(&model.RequestLogin{Account: "alice", Password: "alice"}); errDetail != nil {
		t.Errorf("login with account as password: %+v", errDetail)
	}
}

func TestCreateUserRejectsBlankName(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.CreateUser(&model.RequestCreateUser{Account: "alice", Name: "  ", Role: constant.ROLE_DEFAULT, I18n: constant.I18N_EN})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestListUsersIncludesSystemAdmin(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	resp, errDetail := p.ListUsers()
	if errDetail != nil {
		t.Fatalf("ListUsers: %+v", errDetail)
	}
	if len(resp.Users) != 2 || resp.Users[0].Account != testAdminAccount || resp.Users[1].Account != "alice" {
		t.Errorf("users = %+v", resp.Users)
	}
}

func TestUpdateUserAppliesOnlyGivenFields(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	resp, errDetail := p.UpdateUser("alice", &model.RequestUpdateUser{Role: strPtr(constant.ROLE_ADMIN)})
	if errDetail != nil {
		t.Fatalf("UpdateUser: %+v", errDetail)
	}
	if resp.User.Role != constant.ROLE_ADMIN || resp.User.I18n != constant.I18N_EN {
		t.Errorf("user = %+v", resp.User)
	}
	if _, errDetail := p.Login(&model.RequestLogin{Account: "alice", Password: "alice"}); errDetail != nil {
		t.Errorf("password changed unexpectedly: %+v", errDetail)
	}
}

func TestUpdateUserChangesName(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	resp, errDetail := p.UpdateUser("alice", &model.RequestUpdateUser{Name: strPtr("Alice Wang")})
	if errDetail != nil {
		t.Fatalf("UpdateUser: %+v", errDetail)
	}
	if resp.User.Name != "Alice Wang" || mustGet(t, p, "alice").Name != "Alice Wang" {
		t.Errorf("name not updated: %+v", resp.User)
	}
}

func TestUpdateUserRejectsBlankName(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	_, errDetail := p.UpdateUser("alice", &model.RequestUpdateUser{Name: strPtr(" ")})
	expectStatus(t, errDetail, http.StatusBadRequest)
}

func TestUpdateUserChangesPassword(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	if _, errDetail := p.UpdateUser("alice", &model.RequestUpdateUser{Password: strPtr("reset")}); errDetail != nil {
		t.Fatalf("UpdateUser: %+v", errDetail)
	}
	if _, errDetail := p.Login(&model.RequestLogin{Account: "alice", Password: "reset"}); errDetail != nil {
		t.Errorf("login with reset password: %+v", errDetail)
	}
}

func TestUpdateUserRejectsSystemAdmin(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.UpdateUser(testAdminAccount, &model.RequestUpdateUser{Role: strPtr(constant.ROLE_DEFAULT)})
	expectStatus(t, errDetail, http.StatusForbidden)
}

func TestUpdateUserNotFound(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.UpdateUser("nobody", &model.RequestUpdateUser{Role: strPtr(constant.ROLE_DEFAULT)})
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestDeleteUser(t *testing.T) {
	p := newTestProcessor(t)
	mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	if _, errDetail := p.DeleteUser(mustGet(t, p, testAdminAccount), "alice"); errDetail != nil {
		t.Fatalf("DeleteUser: %+v", errDetail)
	}
	if _, err := p.GetAccount("alice"); err == nil {
		t.Error("account still exists after delete")
	}
}

func TestDeleteUserRejectsSystemAdmin(t *testing.T) {
	p := newTestProcessor(t)
	operator := mustCreateUser(t, p, "bob", constant.ROLE_ADMIN)

	_, errDetail := p.DeleteUser(operator, testAdminAccount)
	expectStatus(t, errDetail, http.StatusForbidden)
}

func TestDeleteUserRejectsSelf(t *testing.T) {
	p := newTestProcessor(t)
	operator := mustCreateUser(t, p, "bob", constant.ROLE_ADMIN)

	_, errDetail := p.DeleteUser(operator, "bob")
	expectStatus(t, errDetail, http.StatusForbidden)
}

func TestDeleteUserNotFound(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.DeleteUser(mustGet(t, p, testAdminAccount), "nobody")
	expectStatus(t, errDetail, http.StatusNotFound)
}

func TestGetMeReturnsUser(t *testing.T) {
	p := newTestProcessor(t)

	resp := p.GetMe(mustGet(t, p, testAdminAccount))
	if resp.User == nil || resp.User.Account != testAdminAccount || !resp.User.IsSystem {
		t.Errorf("user = %+v", resp.User)
	}
}

func TestUpdateMePersistsI18nAndReturnsNewToken(t *testing.T) {
	p := newTestProcessor(t)
	acc := mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	resp, errDetail := p.UpdateMe(acc, &model.RequestUpdateMe{I18n: constant.I18N_ZH_TW})
	if errDetail != nil {
		t.Fatalf("UpdateMe: %+v", errDetail)
	}

	claims, err := util.ValidateJWT(resp.Token, testJwtSecret)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if claims[constant.JWT_CLAIM_I18N] != constant.I18N_ZH_TW {
		t.Errorf("token i18n = %v", claims[constant.JWT_CLAIM_I18N])
	}
	if mustGet(t, p, "alice").I18n != constant.I18N_ZH_TW {
		t.Error("i18n not persisted")
	}
}

func TestChangeMyPassword(t *testing.T) {
	p := newTestProcessor(t)
	acc := mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	if _, errDetail := p.ChangeMyPassword(acc, &model.RequestChangeMyPassword{OldPassword: "alice", NewPassword: "changed"}); errDetail != nil {
		t.Fatalf("ChangeMyPassword: %+v", errDetail)
	}
	if _, errDetail := p.Login(&model.RequestLogin{Account: "alice", Password: "changed"}); errDetail != nil {
		t.Errorf("login with new password: %+v", errDetail)
	}
}

func TestChangeMyPasswordRejectsWrongOldPassword(t *testing.T) {
	p := newTestProcessor(t)
	acc := mustCreateUser(t, p, "alice", constant.ROLE_DEFAULT)

	_, errDetail := p.ChangeMyPassword(acc, &model.RequestChangeMyPassword{OldPassword: "wrong", NewPassword: "changed"})
	expectStatus(t, errDetail, http.StatusForbidden)
}

func TestChangeMyPasswordRejectsSystemAdmin(t *testing.T) {
	p := newTestProcessor(t)

	_, errDetail := p.ChangeMyPassword(mustGet(t, p, testAdminAccount), &model.RequestChangeMyPassword{OldPassword: testAdminPassword, NewPassword: "changed"})
	expectStatus(t, errDetail, http.StatusForbidden)
}
