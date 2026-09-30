package processor

import (
	"backend/constant"
	"backend/model"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func mustCreateApiToken(t *testing.T, p *Processor, acc *model.Account, days int) *model.ResponseCreateApiToken {
	t.Helper()

	resp, errDetail := p.CreateMyApiToken(acc, &model.RequestCreateApiToken{Name: "skill", ExpiresInDays: days})
	if errDetail != nil {
		t.Fatalf("CreateMyApiToken: %+v", errDetail)
	}
	return resp
}

func TestCreateApiTokenStoresOnlyItsHash(t *testing.T) {
	f := newWorkFixture(t)

	resp := mustCreateApiToken(t, f.p, f.alice, 30)

	if !strings.HasPrefix(resp.Token, constant.API_TOKEN_PREFIX) || len(resp.Token) < 40 {
		t.Fatalf("token = %q", resp.Token)
	}
	if !strings.HasPrefix(resp.Token, resp.ApiToken.Prefix) || resp.ApiToken.Name != "skill" {
		t.Errorf("info = %+v", resp.ApiToken)
	}
	if days := resp.ApiToken.ExpiresAt.Sub(resp.ApiToken.CreatedAt).Hours() / 24; days != 30 {
		t.Errorf("lifetime = %v days, want 30", days)
	}

	sum := sha256.Sum256([]byte(resp.Token))
	stored, err := f.p.GetApiToken(hex.EncodeToString(sum[:]))
	if err != nil {
		t.Fatalf("token not stored under its hash: %v", err)
	}
	backup, _ := f.p.Dump()
	for _, token := range backup.ApiTokens {
		if strings.Contains(fmt.Sprintf("%+v", token), resp.Token) {
			t.Error("the plain token is stored")
		}
	}
	if stored.Account != "ALICE" {
		t.Errorf("stored = %+v", stored)
	}
}

func TestCreateApiTokenDefaultsTo365Days(t *testing.T) {
	f := newWorkFixture(t)

	resp := mustCreateApiToken(t, f.p, f.alice, 0)
	if days := resp.ApiToken.ExpiresAt.Sub(resp.ApiToken.CreatedAt).Hours() / 24; days != 365 {
		t.Errorf("lifetime = %v days, want 365", days)
	}
}

func TestCreateApiTokenValidation(t *testing.T) {
	f := newWorkFixture(t)

	_, errDetail := f.p.CreateMyApiToken(f.alice, &model.RequestCreateApiToken{Name: "x", ExpiresInDays: 7})
	expectStatus(t, errDetail, http.StatusBadRequest)
	_, errDetail = f.p.CreateMyApiToken(f.alice, &model.RequestCreateApiToken{Name: "  ", ExpiresInDays: 30})
	expectStatus(t, errDetail, http.StatusBadRequest)

	for i := 0; i < constant.API_TOKEN_MAX_PER_ACCOUNT; i++ {
		mustCreateApiToken(t, f.p, f.alice, 30)
	}
	_, errDetail = f.p.CreateMyApiToken(f.alice, &model.RequestCreateApiToken{Name: "one too many", ExpiresInDays: 30})
	expectStatus(t, errDetail, http.StatusConflict)
}

func TestAuthenticateApiToken(t *testing.T) {
	f := newWorkFixture(t)
	resp := mustCreateApiToken(t, f.p, f.alice, 30)

	acc, errDetail := f.p.AuthenticateApiToken(resp.Token)
	if errDetail != nil || acc.Account != "ALICE" {
		t.Fatalf("AuthenticateApiToken = %+v, %+v", acc, errDetail)
	}

	_, errDetail = f.p.AuthenticateApiToken(constant.API_TOKEN_PREFIX + "made-up")
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestApiTokenLastUsedIsRecordedAtMostPerMinute(t *testing.T) {
	f := newWorkFixture(t)
	resp := mustCreateApiToken(t, f.p, f.alice, 30)
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	f.p.now = func() time.Time { return now }

	lastUsed := func() time.Time {
		list, _ := f.p.ListMyApiTokens(f.alice)
		if list.Tokens[0].LastUsedAt == nil {
			return time.Time{}
		}
		return *list.Tokens[0].LastUsedAt
	}

	if _, errDetail := f.p.AuthenticateApiToken(resp.Token); errDetail != nil {
		t.Fatal(errDetail)
	}
	if !lastUsed().Equal(now) {
		t.Fatalf("lastUsed = %v, want %v", lastUsed(), now)
	}

	first := now
	now = now.Add(30 * time.Second)
	_, _ = f.p.AuthenticateApiToken(resp.Token)
	if !lastUsed().Equal(first) {
		t.Errorf("rewritten within a minute: %v", lastUsed())
	}

	now = now.Add(2 * time.Minute)
	_, _ = f.p.AuthenticateApiToken(resp.Token)
	if !lastUsed().Equal(now) {
		t.Errorf("not recorded after a minute: %v", lastUsed())
	}
}

func TestExpiredApiTokenIsRejected(t *testing.T) {
	f := newWorkFixture(t)
	resp := mustCreateApiToken(t, f.p, f.alice, 30)

	f.p.now = func() time.Time { return resp.ApiToken.ExpiresAt.Add(time.Second) }
	_, errDetail := f.p.AuthenticateApiToken(resp.Token)
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestRevokedApiTokenIsRejected(t *testing.T) {
	f := newWorkFixture(t)
	resp := mustCreateApiToken(t, f.p, f.alice, 30)

	if _, errDetail := f.p.DeleteMyApiToken(f.alice, resp.ApiToken.ID); errDetail != nil {
		t.Fatalf("DeleteMyApiToken: %+v", errDetail)
	}
	_, errDetail := f.p.AuthenticateApiToken(resp.Token)
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestApiTokenOfDeletedAccountIsRejected(t *testing.T) {
	f := newWorkFixture(t)
	resp := mustCreateApiToken(t, f.p, f.alice, 30)

	if _, errDetail := f.p.DeleteUser(mustGet(t, f.p, testAdminAccount), "alice"); errDetail != nil {
		t.Fatal(errDetail)
	}
	_, errDetail := f.p.AuthenticateApiToken(resp.Token)
	expectStatus(t, errDetail, http.StatusUnauthorized)
}

func TestListAndDeleteOnlyOwnApiTokens(t *testing.T) {
	f := newWorkFixture(t)
	bob := mustCreateUser(t, f.p, "bob", constant.ROLE_DEFAULT)
	alices := mustCreateApiToken(t, f.p, f.alice, 30)
	mustCreateApiToken(t, f.p, bob, 30)

	list, errDetail := f.p.ListMyApiTokens(f.alice)
	if errDetail != nil || len(list.Tokens) != 1 || list.Tokens[0].ID != alices.ApiToken.ID {
		t.Errorf("ListMyApiTokens = %+v, %+v", list, errDetail)
	}

	_, errDetail = f.p.DeleteMyApiToken(bob, alices.ApiToken.ID)
	expectStatus(t, errDetail, http.StatusNotFound)
	if _, errDetail := f.p.AuthenticateApiToken(alices.Token); errDetail != nil {
		t.Errorf("someone else revoked the token: %+v", errDetail)
	}
}
