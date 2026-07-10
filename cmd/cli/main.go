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
	"github.com/davidaparicio/namecheck/github"
)

type Result struct {
	Username  string
	Platform  string
	Valid     bool
	Available bool
	Err       error
}

const checkTimeout = 10 * time.Second

func main() {
	if len(os.Args[1:]) == 0 {
		log.Fatal("username args is required")
	}
	username := os.Args[1]

	client := &http.Client{Timeout: checkTimeout}
	checkers := []namecheck.Checker{
		&github.GitHub{Client: client},
	}

	results := make(chan Result, len(checkers))
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	for _, checker := range checkers {
		wg.Add(1)
		go func(checker namecheck.Checker) {
			defer wg.Done()
			results <- check(ctx, checker, username)
		}(checker)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	exitCode := 0
	for res := range results {
		if res.Err != nil {
			exitCode = 1
			var uae *namecheck.UnknownAvailabilityError
			if errors.As(res.Err, &uae) {
				fmt.Fprintf(os.Stderr, "%s: could not check %q: %v\n", uae.Platform, uae.Username, uae.Cause)
				continue
			}
			fmt.Fprintf(os.Stderr, "%s: %v\n", res.Platform, res.Err)
			continue
		}
		fmt.Printf("%s: %q valid=%t available=%t\n", res.Platform, res.Username, res.Valid, res.Available)
	}
	os.Exit(exitCode)
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
	res.Available, res.Err = checker.IsAvailable(ctx, username)
	return res
}
