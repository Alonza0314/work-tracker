package integrationtest

import (
	"net/http"
	"testing"
)

// Admins manage users: create, list, update and delete, with the rules that
// protect the system admin and themselves.
func TestUserManagement(t *testing.T) {
	admin := asAdmin(t)

	t.Run("create stores the account in upper case with the account as password", func(t *testing.T) {
		r := admin.post("/api/users", map[string]string{"account": " bob ", "name": "Bob", "role": "admin", "i18n": "zh-TW"})
		expect(t, r, http.StatusOK, "create bob")
		if user := r.obj("user"); user["account"] != "BOB" || user["name"] != "Bob" || user["role"] != "admin" {
			t.Errorf("user = %v", user)
		}
		login(t, "bob", "bob")
	})

	t.Run("create rejects duplicates and bad input", func(t *testing.T) {
		expect(t, admin.post("/api/users", map[string]string{"account": "Bob", "name": "x", "role": "default", "i18n": "en"}),
			http.StatusConflict, "duplicate account")
		expect(t, admin.post("/api/users", map[string]string{"account": "carl", "name": " ", "role": "default", "i18n": "en"}),
			http.StatusBadRequest, "blank name")
		expect(t, admin.post("/api/users", map[string]string{"account": "carl", "name": "Carl", "role": "root", "i18n": "en"}),
			http.StatusBadRequest, "unknown role")
	})

	t.Run("list is sorted by account", func(t *testing.T) {
		createUser(t, admin, "zoe", "Zoe", "default")
		createUser(t, admin, "aaron", "Aaron", "default")

		users := admin.get("/api/users").list("users")
		if got := ids(users, "account"); got != "AARON,ADMIN,BOB,ZOE" {
			t.Errorf("accounts = %s", got)
		}
		for _, user := range users {
			if (user["account"] == "ADMIN") != (user["isSystem"] == true) {
				t.Errorf("isSystem of %v = %v", user["account"], user["isSystem"])
			}
		}
	})

	t.Run("update changes only the given fields", func(t *testing.T) {
		r := admin.put("/api/users/zoe", map[string]any{"name": "Zoe Wang", "i18n": "zh-TW"})
		expect(t, r, http.StatusOK, "update zoe")
		if user := r.obj("user"); user["name"] != "Zoe Wang" || user["i18n"] != "zh-TW" || user["role"] != "default" {
			t.Errorf("user = %v", user)
		}

		expect(t, admin.put("/api/users/zoe", map[string]any{"password": "Reset1"}), http.StatusOK, "reset password")
		login(t, "zoe", "reset1")

		expect(t, admin.put("/api/users/nobody", map[string]any{"name": "x"}), http.StatusNotFound, "update unknown user")
	})

	t.Run("the system admin cannot be changed or deleted", func(t *testing.T) {
		expect(t, admin.put("/api/users/admin", map[string]any{"role": "default"}), http.StatusForbidden, "update system admin")
		expect(t, admin.del("/api/users/admin"), http.StatusForbidden, "delete system admin")
	})

	t.Run("admins cannot delete themselves", func(t *testing.T) {
		bob := loginAs(t, "bob", "bob")
		expect(t, bob.del("/api/users/bob"), http.StatusForbidden, "bob deletes bob")
	})

	t.Run("default users cannot manage users", func(t *testing.T) {
		aaron := loginAs(t, "aaron", "aaron")
		expect(t, aaron.get("/api/users"), http.StatusForbidden, "list as default user")
		expect(t, aaron.post("/api/users", map[string]string{"account": "x", "name": "x", "role": "default", "i18n": "en"}),
			http.StatusForbidden, "create as default user")
	})

	t.Run("a deleted user's token stops working", func(t *testing.T) {
		zoe := loginAs(t, "zoe", "reset1")
		expect(t, admin.del("/api/users/zoe"), http.StatusOK, "delete zoe")
		expect(t, zoe.get("/api/me"), http.StatusUnauthorized, "zoe after deletion")
		expect(t, admin.del("/api/users/zoe"), http.StatusNotFound, "delete zoe again")
	})
}
