package integrationtest

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// Personal API tokens: created once, used as a bearer credential with the
// account's own permissions, listed without the secret and revoked.
func TestApiTokens(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")
	category := createOption(t, admin, "categories", "Develop")

	create := func(t *testing.T, c *client, body map[string]any) *response {
		t.Helper()
		return c.post("/api/me/api-tokens", body)
	}

	var token, tokenID string

	t.Run("create returns the token once, 365 days by default", func(t *testing.T) {
		r := create(t, alice, map[string]any{"name": "Claude Code skill"})
		expect(t, r, http.StatusOK, "create token")
		token = r.str("token")
		info := r.obj("apiToken")
		tokenID = info["id"].(string)
		if !strings.HasPrefix(token, "wt_") || !strings.HasPrefix(token, info["prefix"].(string)) || info["name"] != "Claude Code skill" {
			t.Errorf("token = %q, info = %v", token, info)
		}
		created, _ := time.Parse(time.RFC3339, info["createdAt"].(string))
		expires, _ := time.Parse(time.RFC3339, info["expiresAt"].(string))
		if days := expires.Sub(created).Hours() / 24; days != 365 {
			t.Errorf("lifetime = %v days", days)
		}
	})

	t.Run("lifetimes are 30, 60, 180 or 365 days", func(t *testing.T) {
		for _, days := range []int{30, 60, 180} {
			expect(t, create(t, admin, map[string]any{"name": "short", "expiresInDays": days}), http.StatusOK, "valid lifetime")
		}
		expect(t, create(t, alice, map[string]any{"name": "x", "expiresInDays": 7}), http.StatusBadRequest, "7 days")
		expect(t, create(t, alice, map[string]any{"name": " "}), http.StatusBadRequest, "blank name")
	})

	t.Run("the token works as a bearer credential", func(t *testing.T) {
		skill := &client{t: t, token: token}
		me := skill.get("/api/me")
		expect(t, me, http.StatusOK, "GET /api/me with the token")
		if me.obj("user")["account"] != "ALICE" {
			t.Errorf("user = %v", me.obj("user"))
		}
		createRecord(t, skill, entry("2026-09-30", category, 1, "logged by a skill"))
		if alice.get("/api/me/work-records?from=2026-09-30&to=2026-09-30").num("total") != 1 {
			t.Error("the record logged with the token is missing")
		}
	})

	t.Run("it has the account's permissions, no more", func(t *testing.T) {
		expect(t, (&client{t: t, token: token}).get("/api/users"), http.StatusForbidden, "admin route with a default user's token")
	})

	t.Run("the list shows the prefix and last use, never the token", func(t *testing.T) {
		r := alice.get("/api/me/api-tokens")
		expect(t, r, http.StatusOK, "list tokens")
		if token == "" || strings.Contains(string(r.body), token) {
			t.Error("the list contains the token")
		}
		tokens := r.list("tokens")
		if len(tokens) != 1 || tokens[0]["id"] != tokenID || tokens[0]["lastUsedAt"] == nil {
			t.Errorf("tokens = %v", tokens)
		}
	})

	t.Run("only the owner revokes a token, and it stops working", func(t *testing.T) {
		expect(t, admin.del("/api/me/api-tokens/"+tokenID), http.StatusNotFound, "someone else revokes")
		expect(t, alice.del("/api/me/api-tokens/"+tokenID), http.StatusOK, "revoke")
		expect(t, (&client{t: t, token: token}).get("/api/me"), http.StatusUnauthorized, "revoked token")
	})

	t.Run("at most 10 tokens per account", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			expect(t, create(t, alice, map[string]any{"name": "t"}), http.StatusOK, "create")
		}
		expect(t, create(t, alice, map[string]any{"name": "one too many"}), http.StatusConflict, "11th token")
	})

	t.Run("deleting the account invalidates its tokens", func(t *testing.T) {
		bob := createUser(t, admin, "bob", "Bob", "default")
		bobToken := create(t, bob, map[string]any{"name": "bob"}).str("token")
		expect(t, admin.del("/api/users/bob"), http.StatusOK, "delete bob")
		expect(t, (&client{t: t, token: bobToken}).get("/api/me"), http.StatusUnauthorized, "token of a deleted account")
	})
}
