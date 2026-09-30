package integrationtest

import (
	"net/http"
	"testing"
)

// The signed-in user's own profile: reading it, the UI language and the
// password.
func TestProfile(t *testing.T) {
	admin := asAdmin(t)
	alice := createUser(t, admin, "alice", "Alice", "default")

	t.Run("get me", func(t *testing.T) {
		r := alice.get("/api/me")
		expect(t, r, http.StatusOK, "GET /api/me")
		user := r.obj("user")
		if user["account"] != "ALICE" || user["name"] != "Alice" || user["role"] != "default" || user["i18n"] != "en" || user["isSystem"] != false {
			t.Errorf("user = %v", user)
		}
	})

	t.Run("changing the language returns a token carrying it", func(t *testing.T) {
		r := alice.put("/api/me", map[string]string{"i18n": "zh-TW"})
		expect(t, r, http.StatusOK, "PUT /api/me")
		if c := claims(t, r.str("token")); c["i18n"] != "zh-TW" {
			t.Errorf("new token i18n = %v", c["i18n"])
		}
		if user := alice.get("/api/me").obj("user"); user["i18n"] != "zh-TW" {
			t.Errorf("stored i18n = %v", user["i18n"])
		}
		expect(t, alice.put("/api/me", map[string]string{"i18n": "fr"}), http.StatusBadRequest, "unsupported language")
	})

	t.Run("change password", func(t *testing.T) {
		wrong := alice.put("/api/me/password", map[string]string{"oldPassword": "wrong", "newPassword": "Secret"})
		expect(t, wrong, http.StatusForbidden, "wrong old password")

		r := alice.put("/api/me/password", map[string]string{"oldPassword": "ALICE", "newPassword": "Secret"})
		expect(t, r, http.StatusOK, "change password")

		login(t, "alice", "secret") // passwords ignore case too
		old := anonymous(t).post("/api/login", map[string]string{"account": "alice", "password": "alice"})
		expect(t, old, http.StatusUnauthorized, "old password")
	})

	t.Run("the system admin password comes from the config", func(t *testing.T) {
		r := admin.put("/api/me/password", map[string]string{"oldPassword": adminPassword, "newPassword": "x"})
		expect(t, r, http.StatusForbidden, "system admin changes password")
	})
}
