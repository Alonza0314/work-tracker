package integrationtest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// the app started by test.sh; override with WT_BASE_URL
func baseURL() string {
	if url := os.Getenv("WT_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:18888"
}

// the system admin of integration-test/config/config.yaml
const (
	adminAccount  = "admin"
	adminPassword = "0000"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// client calls the API with a bearer credential (a JWT or an API token).
type client struct {
	t     *testing.T
	token string
}

// response is a decoded API answer: the status, the raw body and the JSON
// object (nil when the body is not a JSON object).
type response struct {
	status int
	body   []byte
	header http.Header
	json   map[string]any
}

func (r *response) str(key string) string {
	value, _ := r.json[key].(string)
	return value
}

func (r *response) num(key string) float64 {
	value, _ := r.json[key].(float64)
	return value
}

func (r *response) obj(key string) map[string]any {
	value, _ := r.json[key].(map[string]any)
	return value
}

func (r *response) list(key string) []map[string]any {
	items, _ := r.json[key].([]any)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

func (c *client) send(method, path, contentType string, body io.Reader) *response {
	c.t.Helper()

	req, err := http.NewRequest(method, baseURL()+path, body)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("%s %s: reading body: %v", method, path, err)
	}
	result := &response{status: resp.StatusCode, body: data, header: resp.Header}
	_ = json.Unmarshal(data, &result.json)
	return result
}

func (c *client) do(method, path string, payload any) *response {
	c.t.Helper()

	if payload == nil {
		return c.send(method, path, "", nil)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatalf("%s %s: encoding payload: %v", method, path, err)
	}
	return c.send(method, path, "application/json", bytes.NewReader(data))
}

func (c *client) get(path string) *response {
	c.t.Helper()
	return c.do(http.MethodGet, path, nil)
}

func (c *client) post(path string, payload any) *response {
	c.t.Helper()
	return c.do(http.MethodPost, path, payload)
}

func (c *client) put(path string, payload any) *response {
	c.t.Helper()
	return c.do(http.MethodPut, path, payload)
}

func (c *client) del(path string) *response {
	c.t.Helper()
	return c.do(http.MethodDelete, path, nil)
}

// upload posts one file as multipart form data.
func (c *client) upload(path, field, fileName string, content []byte) *response {
	c.t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(field, fileName)
	if err != nil {
		c.t.Fatal(err)
	}
	_, _ = part.Write(content)
	if err := writer.Close(); err != nil {
		c.t.Fatal(err)
	}
	return c.send(http.MethodPost, path, writer.FormDataContentType(), &buf)
}

// expect fails the test unless the response has the wanted status.
func expect(t *testing.T, r *response, want int, what string) {
	t.Helper()

	if r.status != want {
		t.Fatalf("%s: status %d, want %d; body %s", what, r.status, want, r.body)
	}
}

// anonymous is a client without credentials.
func anonymous(t *testing.T) *client {
	return &client{t: t}
}

// login signs in and returns the JWT.
func login(t *testing.T, account, password string) string {
	t.Helper()

	r := anonymous(t).post("/api/login", map[string]string{"account": account, "password": password})
	expect(t, r, http.StatusOK, "login "+account)
	return r.str("token")
}

func loginAs(t *testing.T, account, password string) *client {
	t.Helper()
	return &client{t: t, token: login(t, account, password)}
}

func asAdmin(t *testing.T) *client {
	t.Helper()
	return loginAs(t, adminAccount, adminPassword)
}

// createUser adds a user through the admin API and signs them in (their
// initial password is their account).
func createUser(t *testing.T, admin *client, account, name, role string) *client {
	t.Helper()

	r := admin.post("/api/users", map[string]string{"account": account, "name": name, "role": role, "i18n": "en"})
	expect(t, r, http.StatusOK, "create user "+account)
	return loginAs(t, account, account)
}

// claims decodes the payload of a JWT (without checking its signature).
func claims(t *testing.T, token string) map[string]any {
	t.Helper()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %q", token)
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("JWT payload: %v", err)
	}
	result := map[string]any{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JWT payload: %v", err)
	}
	return result
}

// createOption adds a category or project ("categories" / "projects") and
// returns its ID.
func createOption(t *testing.T, admin *client, kind, name string) string {
	t.Helper()

	r := admin.post("/api/"+kind, map[string]string{"name": name})
	expect(t, r, http.StatusOK, "create "+kind+" "+name)
	return r.obj("option")["id"].(string)
}

// entry is a work record or todo body.
func entry(date, categoryID string, hours float64, description string) map[string]any {
	body := map[string]any{"date": date, "description": description}
	if categoryID != "" {
		body["categoryId"] = categoryID
	}
	if hours != 0 {
		body["hours"] = hours
	}
	return body
}

func createRecord(t *testing.T, c *client, body map[string]any) map[string]any {
	t.Helper()

	r := c.post("/api/me/work-records", body)
	expect(t, r, http.StatusOK, "create work record")
	return r.obj("record")
}

// date returns a YYYY-MM-DD date relative to today (local time, the same
// clock as the container).
func date(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func ids(items []map[string]any, key string) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, fmt.Sprint(item[key]))
	}
	return strings.Join(values, ",")
}
