package integrationtest

import (
	"net/http"
	"strings"
	"testing"
)

// The backend serves the built frontend and falls back to index.html for
// client-side routes.
func TestFrontend(t *testing.T) {
	t.Run("root serves the app", func(t *testing.T) {
		r := anonymous(t).get("/")
		expect(t, r, http.StatusOK, "GET /")
		if !strings.Contains(string(r.body), `<div id="root">`) {
			t.Errorf("GET / is not index.html: %.200s", r.body)
		}
	})

	t.Run("client routes fall back to index.html", func(t *testing.T) {
		for _, path := range []string{"/login", "/work/me", "/users"} {
			r := anonymous(t).get(path)
			expect(t, r, http.StatusOK, "GET "+path)
			if !strings.Contains(string(r.body), `<div id="root">`) {
				t.Errorf("GET %s is not index.html", path)
			}
		}
	})

	t.Run("static files are served", func(t *testing.T) {
		r := anonymous(t).get("/wt-favicon.jpg")
		expect(t, r, http.StatusOK, "GET /wt-favicon.jpg")
		if !strings.HasPrefix(r.header.Get("Content-Type"), "image/") {
			t.Errorf("favicon content type = %q", r.header.Get("Content-Type"))
		}
	})
}
