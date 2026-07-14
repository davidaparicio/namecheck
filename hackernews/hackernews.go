// Package hackernews provides primitives to check if an username
// is available on Hacker News.
package hackernews

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type HackerNews struct {
	Client namecheck.Client
}

const (
	minLen = 2
	maxLen = 15
)

var legalPattern = regexp.MustCompile("^[-0-9A-Z_a-z]*$")

func (*HackerNews) String() string {
	return "HackerNews"
}

func (*HackerNews) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsOnlyLegalChars(username)
}

// IsAvailable checks on Hacker News, the availability of the requested
// username, using the official Firebase-hosted API (no authentication
// required). The endpoint always answers 200; a literal "null" body
// indicates the username's availability.
func (hn *HackerNews) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/user/%s.json", url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: hn.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := hn.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: hn.String(),
			Cause:    err,
		}
		return false, &err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: hn.String(),
			Cause:    fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
		return false, &err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: hn.String(),
			Cause:    err,
		}
		return false, &err
	}
	return strings.TrimSpace(string(bodyBytes)) == "null", nil
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}
