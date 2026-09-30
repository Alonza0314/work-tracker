package integrationtest

import (
	"net/http"
	"testing"
)

// Signing in with the config admin, and what a missing or bad credential gets.
func TestLogin(t *testing.T) {
	t.Run("system admin signs in with a JWT", func(t *testing.T) {
		token := login(t, adminAccount, adminPassword)
		c := claims(t, token)
		if c["sub"] != "ADMIN" || c["role"] != "admin" || c["i18n"] != "zh-TW" || c["name"] != "ADMIN" {
			t.Errorf("claims = %v", c)
		}
	})

	t.Run("account and password ignore case", func(t *testing.T) {
		for _, account := range []string{"Admin", "ADMIN", " admin "} {
			login(t, account, adminPassword)
		}
	})

	t.Run("wrong password or unknown account is 401", func(t *testing.T) {
		for _, body := range []map[string]string{
			{"account": adminAccount, "password": "wrong"},
			{"account": "nobody", "password": "x"},
		} {
			r := anonymous(t).post("/api/login", body)
			expect(t, r, http.StatusUnauthorized, "login "+body["account"])
		}
	})

	t.Run("missing fields are 400", func(t *testing.T) {
		r := anonymous(t).post("/api/login", map[string]string{"account": adminAccount})
		expect(t, r, http.StatusBadRequest, "login without password")
	})

	t.Run("protected routes need a valid bearer credential", func(t *testing.T) {
		expect(t, anonymous(t).get("/api/me"), http.StatusUnauthorized, "no token")
		expect(t, (&client{t: t, token: "not.a.jwt"}).get("/api/me"), http.StatusUnauthorized, "bad JWT")
		expect(t, (&client{t: t, token: "wt_unknown"}).get("/api/me"), http.StatusUnauthorized, "unknown API token")
	})

	t.Run("logout answers 204", func(t *testing.T) {
		expect(t, asAdmin(t).post("/api/logout", nil), http.StatusNoContent, "logout")
	})
}
