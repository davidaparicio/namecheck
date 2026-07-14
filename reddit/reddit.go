// Package reddit provides primitives to check if an username
// is available on Reddit.
package reddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type Reddit struct {
	Client namecheck.Client
}

const (
	minLen = 3
	maxLen = 20
)

var legalPattern = regexp.MustCompile("^[-0-9A-Z_a-z]*$")

func (*Reddit) String() string {
	return "Reddit"
}

func (*Reddit) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsOnlyLegalChars(username)
}

// IsAvailable checks on Reddit, the availability of the requested username.
// A 404 status code on the user's about page indicates the username's
// availability.
func (red *Reddit) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://www.reddit.com/user/%s/about.json", url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: red.String(),
			Cause:    err,
		}
		return false, &err
	}
	// Reddit rejects requests without a User-Agent header.
	req.Header.Set("User-Agent", "namecheck/1.0 (https://github.com/davidaparicio/namecheck)")
	resp, err := red.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: red.String(),
			Cause:    err,
		}
		return false, &err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return true, nil
	case http.StatusOK:
		return false, nil
	default:
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: red.String(),
			Cause:    fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
		return false, &err
	}
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}
