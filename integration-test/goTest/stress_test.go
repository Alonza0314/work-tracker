package integrationtest

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// the load of TestStress; override with WT_STRESS_USERS / WT_STRESS_RECORDS
func stressSize(t *testing.T, name string, fallback int) int {
	t.Helper()

	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 2 {
		t.Fatalf("%s=%q: want a number of at least 2", name, value)
	}
	return n
}

// stressStats collects the outcome of concurrent requests: unexpected answers
// become errors (reported once at the end) and every call's latency is kept.
type stressStats struct {
	mu        sync.Mutex
	errors    []string
	latencies []time.Duration
}

// call runs one request and checks its status; it returns nil when the
// request failed or answered another status.
func (s *stressStats) call(c *client, method, path string, payload any, want int) *response {
	start := time.Now()
	r, err := c.try(method, path, payload)
	elapsed := time.Since(start)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.latencies = append(s.latencies, elapsed)
	switch {
	case err != nil:
		s.errors = append(s.errors, err.Error())
		return nil
	case r.status != want:
		s.errors = append(s.errors, fmt.Sprintf("%s %s: status %d, want %d; body %s", method, path, r.status, want, r.body))
		return nil
	}
	return r
}

func (s *stressStats) fail(format string, args ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = append(s.errors, fmt.Sprintf(format, args...))
}

// check fails the test with the first errors and logs the latencies.
func (s *stressStats) check(t *testing.T, what string, elapsed time.Duration) {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.latencies); n > 0 {
		sorted := append([]time.Duration(nil), s.latencies...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
		t.Logf("%s: %d requests in %v (%.0f req/s), p50 %v, p95 %v, max %v", what, n,
			elapsed.Round(time.Millisecond), float64(n)/elapsed.Seconds(),
			sorted[n/2], sorted[n*95/100], sorted[n-1])
	}
	for i, message := range s.errors {
		if i == 10 {
			t.Errorf("... and %d more", len(s.errors)-i)
			break
		}
		t.Error(message)
	}
}

// parallel runs fn(0..n-1) on n goroutines started together.
func parallel(n int, fn func(i int)) time.Duration {
	var ready, done sync.WaitGroup
	start := make(chan struct{})
	ready.Add(n)
	done.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			fn(i)
		}()
	}
	ready.Wait()
	began := time.Now()
	close(start)
	done.Wait()
	return time.Since(began)
}

// Many users writing and reading at the same time: no request fails, no write
// is lost, every read is a consistent snapshot, and racing completions of one
// todo create exactly one record.
func TestStress(t *testing.T) {
	users := stressSize(t, "WT_STRESS_USERS", 10)
	perUser := stressSize(t, "WT_STRESS_RECORDS", 40)

	admin := asAdmin(t)
	if r := admin.put("/api/settings/work", map[string]any{"allowViewAll": true}); r.status != http.StatusOK {
		t.Fatalf("allowViewAll: status %d", r.status)
	}
	category := createOption(t, admin, "categories", "Develop")
	members := make([]*client, users)
	for i := range members {
		account := fmt.Sprintf("stress%02d", i)
		members[i] = createUser(t, admin, account, "Stress "+account, "default")
	}

	// the records of one user: one per day of 2026 (wrapping), 0.5 to 4 hours
	firstDay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	recordDate := func(j int) string { return firstDay.AddDate(0, 0, j%365).Format("2006-01-02") }
	recordHours := func(j int) float64 { return float64(1+j%8) / 2 }
	var wantHours float64
	for j := 0; j < perUser; j++ {
		wantHours += recordHours(j)
	}
	const everyone = "/api/work-records?from=2026-01-01&to=2026-12-31"
	recordIDs := make([][]string, users)

	t.Run("concurrent writers and readers", func(t *testing.T) {
		stats := &stressStats{}
		var writing atomic.Int32
		writing.Store(int32(users))

		// one goroutine per member writes, three more read everyone's table
		readers := min(3, users)
		elapsed := parallel(users+readers, func(i int) {
			if i >= users {
				reader := members[i-users]
				seen := 0
				for writing.Load() > 0 {
					r := stats.call(reader, http.MethodGet, everyone, nil, http.StatusOK)
					if r == nil {
						return
					}
					var hours float64
					for _, record := range r.list("records") {
						hours += record["hours"].(float64)
					}
					total := int(r.num("total"))
					if total != len(r.list("records")) || hours != r.num("totalHours") {
						stats.fail("inconsistent read: total %d with %d records, totalHours %v vs %v", total, len(r.list("records")), r.num("totalHours"), hours)
					}
					if total < seen {
						stats.fail("records went from %d back to %d", seen, total)
					}
					seen = total
				}
				return
			}

			defer writing.Add(-1)
			for j := 0; j < perUser; j++ {
				r := stats.call(members[i], http.MethodPost, "/api/me/work-records",
					entry(recordDate(j), category, recordHours(j), fmt.Sprintf("user %d record %d", i, j)), http.StatusOK)
				if r != nil {
					recordIDs[i] = append(recordIDs[i], r.obj("record")["id"].(string))
				}
			}
		})
		stats.check(t, "write + read", elapsed)

		r := admin.get(everyone)
		expect(t, r, http.StatusOK, "everyone's records")
		if got, want := int(r.num("total")), users*perUser; got != want {
			t.Fatalf("total = %d, want %d", got, want)
		}
		if got, want := r.num("totalHours"), wantHours*float64(users); got != want {
			t.Errorf("totalHours = %v, want %v", got, want)
		}
		unique := map[string]bool{}
		for _, record := range r.list("records") {
			unique[record["id"].(string)] = true
		}
		if len(unique) != users*perUser {
			t.Errorf("%d unique IDs, want %d", len(unique), users*perUser)
		}
		for i, member := range members {
			own := member.get("/api/me/work-records?from=2026-01-01&to=2026-12-31")
			if int(own.num("total")) != perUser || own.num("totalHours") != wantHours {
				t.Errorf("user %d: total %v, hours %v", i, own.num("total"), own.num("totalHours"))
			}
		}
	})

	t.Run("concurrent updates and deletes", func(t *testing.T) {
		stats := &stressStats{}
		elapsed := parallel(users, func(i int) {
			for j, id := range recordIDs[i] {
				path := "/api/me/work-records/" + id
				if j%2 == 0 {
					stats.call(members[i], http.MethodDelete, path, nil, http.StatusOK)
				} else {
					stats.call(members[i], http.MethodPut, path, entry(recordDate(j), category, 2, "updated"), http.StatusOK)
				}
			}
		})
		stats.check(t, "update + delete", elapsed)

		kept := perUser / 2
		r := admin.get(everyone)
		if got := int(r.num("total")); got != users*kept {
			t.Errorf("total = %d, want %d", got, users*kept)
		}
		if got := r.num("totalHours"); got != float64(users*kept*2) {
			t.Errorf("totalHours = %v, want %d", got, users*kept*2)
		}
	})

	t.Run("racing completions of one todo create one record", func(t *testing.T) {
		alice := members[0]
		before := int(alice.get("/api/me/work-records?from=2026-01-01&to=2026-12-31").num("total"))
		r := alice.post("/api/me/todos", entry("2026-11-02", category, 1, "only once"))
		expect(t, r, http.StatusOK, "create todo")
		path := "/api/me/todos/" + r.obj("todo")["id"].(string) + "/complete"

		var completed, missing atomic.Int32
		stats := &stressStats{}
		elapsed := parallel(20, func(int) {
			r, err := alice.try(http.MethodPost, path, map[string]any{"date": "2026-11-02"})
			switch {
			case err != nil:
				stats.fail("%v", err)
			case r.status == http.StatusOK:
				completed.Add(1)
			case r.status == http.StatusNotFound:
				missing.Add(1)
			default:
				stats.fail("complete: status %d; body %s", r.status, r.body)
			}
		})
		stats.check(t, "complete", elapsed)
		if completed.Load() != 1 || missing.Load() != 19 {
			t.Errorf("completed %d times, 404 %d times; want 1 and 19", completed.Load(), missing.Load())
		}
		after := int(alice.get("/api/me/work-records?from=2026-01-01&to=2026-12-31").num("total"))
		if after != before+1 {
			t.Errorf("records %d -> %d, want one more", before, after)
		}
	})

	t.Run("concurrent logins and API token calls", func(t *testing.T) {
		r := members[1].post("/api/me/api-tokens", map[string]any{"name": "stress"})
		expect(t, r, http.StatusOK, "create API token")
		bot := &client{t: t, token: r.str("token")}

		stats := &stressStats{}
		elapsed := parallel(users*2, func(i int) {
			if i < users {
				account := fmt.Sprintf("STRESS%02d", i)
				stats.call(anonymous(t), http.MethodPost, "/api/login",
					map[string]string{"account": account, "password": account}, http.StatusOK)
				return
			}
			for range 10 {
				stats.call(bot, http.MethodGet, "/api/me", nil, http.StatusOK)
			}
		})
		stats.check(t, "login + token", elapsed)
	})

	t.Run("backup while writing", func(t *testing.T) {
		stats := &stressStats{}
		var backup *response
		elapsed := parallel(users+1, func(i int) {
			if i == users {
				backup = stats.call(admin, http.MethodGet, "/api/system/backup", nil, http.StatusOK)
				return
			}
			for j := 0; j < 5; j++ {
				stats.call(members[i], http.MethodPost, "/api/me/work-records",
					entry("2026-12-01", category, 1, "during backup"), http.StatusOK)
			}
		})
		stats.check(t, "backup + write", elapsed)
		if backup == nil {
			return
		}

		reader, err := zip.NewReader(bytes.NewReader(backup.body), int64(len(backup.body)))
		if err != nil {
			t.Fatalf("backup is not a zip: %v", err)
		}
		for _, file := range reader.File {
			if file.Name != "work.json" {
				continue
			}
			f, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			var records []map[string]any
			err = json.NewDecoder(f).Decode(&records)
			_ = f.Close()
			if err != nil {
				t.Fatalf("work.json: %v", err)
			}
			// a snapshot: everything written before, part of what was written during
			low, high := users*(perUser/2)+1, users*(perUser/2)+1+users*5
			if len(records) < low || len(records) > high {
				t.Errorf("backup has %d records, want %d..%d", len(records), low, high)
			}
			return
		}
		t.Error("backup has no work.json")
	})
}
