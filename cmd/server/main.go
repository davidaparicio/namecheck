package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	//Be careful!!! http://mmcloughlin.com/posts/your-pprof-is-showing
	//_ "net/http/pprof"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/github"
	"github.com/davidaparicio/namecheck/tinder"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

type Result struct {
	Username  string
	Platform  string
	Valid     bool
	Available bool
	Error     error
}

const (
	// maxOutstanding bounds the number of concurrent availability checks.
	maxOutstanding = 16
	// maxUsernameLen is the longest username any supported platform allows.
	maxUsernameLen = 39
	// maxTrackedNames caps the size of the per-username visit map so that
	// attacker-chosen usernames cannot grow it without bound.
	maxTrackedNames = 10_000
	// outboundTimeout bounds each availability check against a third party.
	outboundTimeout = 10 * time.Second
)

type server struct {
	checkers []namecheck.Checker

	visits atomic.Uint64

	mu     sync.Mutex
	counts map[string]uint

	// statsToken guards /stats and /details, which expose every username
	// ever queried. If empty, those endpoints are disabled.
	statsToken string
}

func main() {
	// Clients and Transports are safe for concurrent use by multiple
	// goroutines, and for efficiency should only be created once and reused.
	client := &http.Client{Timeout: outboundTimeout}
	s := &server{
		checkers: []namecheck.Checker{
			&github.GitHub{Client: client},
			&tinder.Tinder{Client: client},
		},
		counts:     make(map[string]uint),
		statsToken: os.Getenv("NAMECHECK_STATS_TOKEN"),
	}

	r := mux.NewRouter()
	r.HandleFunc("/check", s.handleCheck)
	r.HandleFunc("/stats", s.handleStats)
	r.HandleFunc("/visits", s.handleVisits)
	r.HandleFunc("/details", s.handleDetails)

	limiter := newIPRateLimiter(rate.Limit(10), 20)

	srv := &http.Server{
		Addr:        ":8080",
		ReadTimeout: 5 * time.Second,
		// WriteTimeout must exceed outboundTimeout: /check only starts
		// writing once the upstream availability checks have finished.
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           limiter.middleware(r),
		//TLSConfig:       tlsConfig,
	}

	log.Println("Server running on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// authorize allows access to sensitive endpoints only when a stats token is
// configured and the caller presents it as a bearer token.
func (s *server) authorize(w http.ResponseWriter, r *http.Request) bool {
	if s.statsToken == "" {
		http.Error(w, "stats endpoints are disabled (set NAMECHECK_STATS_TOKEN to enable)", http.StatusForbidden)
		return false
	}
	want := "Bearer " + s.statsToken
	got := r.Header.Get("Authorization")
	if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := fmt.Fprint(w, s.counts); err != nil {
		log.Printf("writing stats response: %v", err)
	}
}

func (s *server) handleDetails(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := enc.Encode(s.counts); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *server) handleVisits(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	entity := struct {
		Count uint64 `json:"visits"`
	}{
		Count: s.visits.Load(),
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(entity); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *server) handleCheck(w http.ResponseWriter, r *http.Request) {
	s.visits.Add(1)
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "'username' query param is required", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(username) > maxUsernameLen {
		http.Error(w, "'username' query param is too long", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	// Only track new usernames while the map is below its cap; known
	// usernames keep counting.
	if _, seen := s.counts[username]; seen || len(s.counts) < maxTrackedNames {
		s.counts[username]++
	}
	s.mu.Unlock()

	results := make(chan Result)
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxOutstanding)
	for _, checker := range s.checkers {
		wg.Add(1)
		go func(checker namecheck.Checker) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results <- check(r.Context(), checker, username)
		}(checker)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	w.Header().Set("Content-Type", "application/json")
	type jsonResult struct {
		Platform  string `json:"platform"`
		Valid     string `json:"valid"`
		Available string `json:"available"`
	}
	jsonResults := make([]jsonResult, 0, len(s.checkers))
	for result := range results {
		res := jsonResult{
			Platform:  result.Platform,
			Valid:     fmt.Sprintf("%t", result.Valid),
			Available: availabilityStatus(result),
		}
		jsonResults = append(jsonResults, res)
	}
	entity := struct {
		Username string       `json:"username"`
		Results  []jsonResult `json:"results"`
	}{
		Username: username,
		Results:  jsonResults,
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(entity); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func check(ctx context.Context, checker namecheck.Checker, username string) Result {
	res := Result{
		Username: username,
		Platform: checker.String(),
	}
	res.Valid = checker.IsValid(username)
	if !res.Valid {
		return res
	}
	res.Available, res.Error = checker.IsAvailable(ctx, username)
	return res
}

func availabilityStatus(res Result) string {
	if res.Error != nil {
		return "unknown"
	}
	return fmt.Sprintf("%t", res.Available)
}

// ipRateLimiter applies a per-client-IP token-bucket rate limit.
type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(limit rate.Limit, burst int) *ipRateLimiter {
	rl := &ipRateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		burst:    burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func (rl *ipRateLimiter) cleanupLoop() {
	const staleAfter = 3 * time.Minute
	for range time.Tick(time.Minute) {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > staleAfter {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *ipRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.allow(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
