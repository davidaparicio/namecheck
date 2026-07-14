package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/bluesky"
	"github.com/davidaparicio/namecheck/github"
	"github.com/davidaparicio/namecheck/gitlab"
	"github.com/davidaparicio/namecheck/hackernews"
	"github.com/davidaparicio/namecheck/mastodon"
	"github.com/davidaparicio/namecheck/reddit"
)

//type Status int

type Result struct {
	Username  string
	Platform  string
	Valid     bool
	Available bool
}

/*const (
	Unknown Status = iota
	Active
	Suspended
	Available
)*/

func main() {
	if len(os.Args[1:]) == 0 {
		log.Fatal("username args is required")
	}
	username := os.Args[1]

	/*t := &twitter.Twitter{
		Client: http.DefaultClient,
	}*/
	checkers := []namecheck.Checker{
		&github.GitHub{Client: http.DefaultClient},
		&gitlab.GitLab{Client: http.DefaultClient},
		&reddit.Reddit{Client: http.DefaultClient},
		&bluesky.Bluesky{Client: http.DefaultClient},
		&mastodon.Mastodon{Client: http.DefaultClient},
		&hackernews.HackerNews{Client: http.DefaultClient},
	}
	results := make(chan Result, len(checkers))
	errc := make(chan error, len(checkers))
	var wg sync.WaitGroup
	wg.Add(len(checkers))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, checker := range checkers {
		go check(ctx, checker, username, &wg, results, errc)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	// each checker sends exactly one message, either a result or an error
	for i := 0; i < len(checkers); i++ {
		select {
		case err := <-errc:
			//OLD
			//const tmpl = "namecheck: some error occurred: %s\n"
			//fmt.Fprintf(os.Stderr, tmpl, err)
			//NEW since the Go 1.13
			/*type wrapper interface{ Unwrap() error }
			if err, ok := err.(wrapper); ok { // err has a cause
				// call err.Unwrap to access the error that caused err
				fmt.Println(err.Unwrap())
			}*/
			var uae *namecheck.UnknownAvailabilityError
			// Errors.Is for default errrors
			if errors.As(err, &uae) {
				fmt.Println(uae.Platform, uae.Username)
			}
			fmt.Fprintln(os.Stderr, err)
		case res := <-results:
			fmt.Println(res)
		}
	}
}

func check(
	ctx context.Context,
	checker namecheck.Checker,
	username string,
	wg *sync.WaitGroup,
	results chan<- Result,
	errc chan<- error,
) {
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
	if err != nil {
		errc <- err
		return
	}
	res.Available = avail
	results <- res
}
