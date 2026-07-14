// Package gitlab provides primitives to check if an username
// is available on GitLab.
package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type GitLab struct {
	Client namecheck.Client
}

const (
	minLen         = 2
	maxLen         = 255
	illegalPrefix  = "-"
	illegalPattern = "--"
)

var legalPattern = regexp.MustCompile("^[0-9A-Za-z][-0-9A-Za-z_.]*$")
var illegalSuffixes = []string{"-", ".", ".git", ".atom"}

func (*GitLab) String() string {
	return "GitLab"
}

func (*GitLab) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen) &&
		containsNoIllegalPattern(username) &&
		containsOnlyLegalChars(username) &&
		containsNoIllegalPrefix(username) &&
		containsNoIllegalSuffix(username)
}

// IsAvailable checks on GitLab, the availability of the requested username,
// using the public REST API (no authentication required).
// An empty result set indicates the username's availability.
func (gl *GitLab) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://gitlab.com/api/v4/users?username=%s", url.QueryEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: gl.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := gl.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: gl.String(),
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
			Platform: gl.String(),
			Cause:    fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
		return false, &err
	}

	var users []json.RawMessage
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&users); err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: gl.String(),
			Cause:    err,
		}
		return false, &err
	}
	return len(users) == 0, nil
}

func containsNoIllegalPattern(username string) bool {
	return !strings.Contains(username, illegalPattern)
}

func containsOnlyLegalChars(username string) bool {
	return legalPattern.MatchString(username)
}

func containsNoIllegalPrefix(username string) bool {
	return !strings.HasPrefix(username, illegalPrefix)
}

func containsNoIllegalSuffix(username string) bool {
	for _, suffix := range illegalSuffixes {
		if strings.HasSuffix(username, suffix) {
			return false
		}
	}
	return true
}
