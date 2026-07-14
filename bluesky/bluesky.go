// Package bluesky provides primitives to check if an username
// (a *.bsky.social handle) is available on Bluesky.
package bluesky

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type Bluesky struct {
	Client namecheck.Client
}

const (
	minLen        = 3
	maxLen        = 18
	illegalPrefix = "-"
	illegalSuffix = "-"
)

var legalPattern = regexp.MustCompile("^[-0-9A-Za-z]*$")

func (*Bluesky) String() string {
	return "Bluesky"
}

func (*Bluesky) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsOnlyLegalChars(username) &&
		containsNoIllegalPrefix(username) &&
		containsNoIllegalSuffix(username)
}

// IsAvailable checks on Bluesky, the availability of the requested username
// as a handle under the default bsky.social domain, using the public
// AT Protocol API (no authentication required).
// A 400 status code (unresolvable handle) indicates the username's
// availability.
func (bs *Bluesky) IsAvailable(ctx context.Context, username string) (bool, error) {
	handle := strings.ToLower(username) + ".bsky.social"
	endpoint := fmt.Sprintf(
		"https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=%s",
		url.QueryEscape(handle),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: bs.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := bs.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: bs.String(),
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
	case http.StatusBadRequest:
		return true, nil
	case http.StatusOK:
		return false, nil
	default:
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: bs.String(),
			Cause:    fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
		return false, &err
	}
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}

func containsNoIllegalPrefix(username string) bool {
	return !strings.HasPrefix(username, illegalPrefix)
}

func containsNoIllegalSuffix(username string) bool {
	return !strings.HasSuffix(username, illegalSuffix)
}
