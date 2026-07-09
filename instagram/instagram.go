// Package instagram provides primitives to check if an username
// is available on Instagram.
package instagram

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

type Instagram struct {
	Client namecheck.Client
}

const (
	minLen = 3
	maxLen = 30
)

// Instagram usernames may contain letters, digits, periods and underscores.
var legalPattern = regexp.MustCompile("^[0-9A-Za-z._]*$")

func (*Instagram) String() string {
	return "Instagram"
}

func (*Instagram) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsOnlyLegalChars(username)
}

func (insta *Instagram) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://www.instagram.com/%s/?locale=en_US", url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: insta.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := insta.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: insta.String(),
			Cause:    err,
		}
		return false, &err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: insta.String(),
			Cause:    fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		}
		return false, &err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: insta.String(),
			Cause:    fmt.Errorf("error reading response body: %w", err),
		}
		return false, &err
	}
	bodyString := string(bodyBytes)

	/* if !noimageindex  = pseudo AVAILABLE (True)
	   else noimageindex = NOT AVAILABLE (False) */
	return !strings.Contains(bodyString, "noarchive, noimageindex"), nil
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}
