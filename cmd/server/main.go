package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	//Be careful!!! http://mmcloughlin.com/posts/your-pprof-is-showing
	//_ "net/http/pprof"

	"sync"
	"sync/atomic"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/github"
	"github.com/gorilla/mux"
)

//type Status int

type Result struct {
	Username  string
	Platform  string
	Valid     bool
	Available bool
	Error     error
}

/*const (
	Unknown Status = iota
	Active
	Suspended
	Available
)*/

var (
	visits uint64
	m      = make(map[string]uint)
	mu     sync.Mutex
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/check", handleCheck)
	r.HandleFunc("/stats", handleStats)
	r.HandleFunc("/visits", handleVisits)
	r.HandleFunc("/details", handleDetails)
	http.Handle("/", r)

	srv := &http.Server{
		Addr:              ":8080",
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           r,
		//TLSConfig:       tlsConfig,
	}

	log.Println("Server running on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleStats(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	fmt.Fprint(w, m)
	mu.Unlock()
}

func handleDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dec := json.NewEncoder(w)
	mu.Lock()
	defer mu.Unlock()
	if err := dec.Encode(m); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func handleVisits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	entity := struct {
		Count uint64 `json:"visits"`
	}{
		Count: atomic.LoadUint64(&visits),
	}
	dec := json.NewEncoder(w)
	if err := dec.Encode(entity); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	atomic.AddUint64(&visits, 1)
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "'username' query param is required", http.StatusBadRequest)
		return
	}
	mu.Lock()
	m[username]++
	mu.Unlock()
	var checkers []namecheck.Checker
	//for i := 0; i < 50; i++ {
	for i := 0; i < 3; i++ {
		//Clients and Transports are safe for concurrent use by multiple
		//goroutines and for efficiency should only be created once and re-used.
		//So no DATA RACE ;)
		/*t := &twitter.Twitter{
			Client: http.DefaultClient,
		}*/
		g := &github.GitHub{
			Client: http.DefaultClient,
		}
		checkers = append(checkers, g)
	}
	results := make(chan Result)
	/*ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()*/
	var wg sync.WaitGroup
	wg.Add(len(checkers))
	const maxOutstanding = 16
	sem := make(chan struct{}, maxOutstanding)
	for _, checker := range checkers {
		go check(r.Context(), checker, username, &wg, sem, results)
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
	jsonResults := make([]jsonResult, 0, len(checkers))
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
	dec := json.NewEncoder(w)
	if err := dec.Encode(entity); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func check(
	ctx context.Context,
	checker namecheck.Checker,
	username string,
	wg *sync.WaitGroup,
	sem <-chan struct{},
	results chan<- Result,
) {
	defer func() { <-sem }()
	defer wg.Done()
	res := Result{
		Username: username,
		Platform: checker.String(),
	}
	res.Valid = checker.IsValid(username)
	if !res.Valid {
		results <- res
		return
	}
	avail, err := checker.IsAvailable(ctx, username)
	res.Available = avail
	if err != nil {
		res.Error = err
	}
	results <- res
}

func availabilityStatus(res Result) string {
	if res.Error != nil {
		return "unknown"
	}
	return fmt.Sprintf("%t", res.Available)
}
