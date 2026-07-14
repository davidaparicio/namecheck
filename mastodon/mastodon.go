// Package mastodon provides primitives to check if an username
// is available on Mastodon (on the mastodon.social instance).
package mastodon

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type Mastodon struct {
	Client namecheck.Client
}

const (
	minLen = 1
	maxLen = 30
)

var legalPattern = regexp.MustCompile("^[0-9A-Z_a-z]*$")

func (*Mastodon) String() string {
	return "Mastodon"
}

func (*Mastodon) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsOnlyLegalChars(username)
}

// IsAvailable checks on Mastodon (mastodon.social instance),
// the availability of the requested username, using the public REST API
// (no authentication required).
// A 404 status code on the account lookup indicates the username's
// availability.
func (mast *Mastodon) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://mastodon.social/api/v1/accounts/lookup?acct=%s", url.QueryEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: mast.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := mast.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: mast.String(),
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
			Platform: mast.String(),
			Cause:    fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
		return false, &err
	}
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}
