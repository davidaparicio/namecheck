package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/falser"
	"github.com/davidaparicio/namecheck/truer"
)

type errChecker struct{}

func (errChecker) IsValid(string) bool { return true }

func (errChecker) IsAvailable(context.Context, string) (bool, error) {
	return false, errors.New("boom")
}

func (errChecker) String() string { return "Erroring" }

func newTestServer(checkers ...namecheck.Checker) *server {
	return &server{
		checkers:   checkers,
		counts:     make(map[string]uint),
		statsToken: "s3cret",
	}
}

type checkResponse struct {
	Username string `json:"username"`
	Results  []struct {
		Platform  string `json:"platform"`
		Valid     string `json:"valid"`
		Available string `json:"available"`
	} `json:"results"`
}

func TestHandleCheck(t *testing.T) {
	s := newTestServer(&truer.Truer{}, &falser.Falser{}, errChecker{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check?username=whatever", nil)
	s.handleCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rec.Code, http.StatusOK)
	}
	var resp checkResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Username != "whatever" {
		t.Errorf("username = %q; want %q", resp.Username, "whatever")
	}
	if len(resp.Results) != 3 {
		t.Fatalf("got %d results; want 3", len(resp.Results))
	}
	want := map[string][2]string{
		"Truer":    {"true", "true"},
		"Falser":   {"false", "false"},
		"Erroring": {"true", "unknown"},
	}
	for _, res := range resp.Results {
		expected, ok := want[res.Platform]
		if !ok {
			t.Errorf("unexpected platform %q", res.Platform)
			continue
		}
		if res.Valid != expected[0] || res.Available != expected[1] {
			t.Errorf(
				"%s: valid=%s available=%s; want valid=%s available=%s",
				res.Platform, res.Valid, res.Available, expected[0], expected[1],
			)
		}
	}
}

func TestHandleCheckMissingUsername(t *testing.T) {
	s := newTestServer(&truer.Truer{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	s.handleCheck(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleCheckUsernameTooLong(t *testing.T) {
	s := newTestServer(&truer.Truer{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check?username="+strings.Repeat("a", maxUsernameLen+1), nil)
	s.handleCheck(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestVisitCountsAreCapped(t *testing.T) {
	s := newTestServer(&truer.Truer{})
	for i := 0; i < maxTrackedNames; i++ {
		s.counts["existing"+string(rune(i))] = 1
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check?username=newcomer", nil)
	s.handleCheck(rec, req)
	if _, tracked := s.counts["newcomer"]; tracked {
		t.Error("a new username was tracked beyond the map cap")
	}
	if len(s.counts) != maxTrackedNames {
		t.Errorf("len(counts) = %d; want %d", len(s.counts), maxTrackedNames)
	}
}

func TestHandleVisits(t *testing.T) {
	s := newTestServer(&truer.Truer{})
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/check?username=whatever", nil)
		s.handleCheck(rec, req)
	}
	rec := httptest.NewRecorder()
	s.handleVisits(rec, httptest.NewRequest(http.MethodGet, "/visits", nil))
	var entity struct {
		Count uint64 `json:"visits"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&entity); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if entity.Count != 3 {
		t.Errorf("visits = %d; want 3", entity.Count)
	}
}

func TestStatsEndpointsRequireToken(t *testing.T) {
	handlers := map[string]func(*server) http.HandlerFunc{
		"/stats":   func(s *server) http.HandlerFunc { return s.handleStats },
		"/details": func(s *server) http.HandlerFunc { return s.handleDetails },
	}
	for path, handler := range handlers {
		t.Run(path, func(t *testing.T) {
			s := newTestServer(&truer.Truer{})

			rec := httptest.NewRecorder()
			handler(s)(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("no token: status = %d; want %d", rec.Code, http.StatusUnauthorized)
			}

			rec = httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer wrong")
			handler(s)(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("wrong token: status = %d; want %d", rec.Code, http.StatusUnauthorized)
			}

			rec = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer s3cret")
			handler(s)(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("valid token: status = %d; want %d", rec.Code, http.StatusOK)
			}

			s.statsToken = ""
			rec = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer s3cret")
			handler(s)(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Errorf("disabled: status = %d; want %d", rec.Code, http.StatusForbidden)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	rl := newIPRateLimiter(1, 2)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rl.middleware(next)

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/check?username=whatever", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		h.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
		t.Errorf("first two requests = %v; want them allowed", codes[:2])
	}
	if codes[2] != http.StatusTooManyRequests {
		t.Errorf("third request = %d; want %d", codes[2], http.StatusTooManyRequests)
	}

	// A different IP has its own bucket.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check?username=whatever", nil)
	req.RemoteAddr = "192.0.2.2:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("other IP: status = %d; want %d", rec.Code, http.StatusOK)
	}
}
