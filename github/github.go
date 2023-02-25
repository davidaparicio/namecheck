// Package github provides primitives to check if an username
// is available on Github.
package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/internal"
)

type GitHub struct {
	Client namecheck.Client
}

const (
	minLen         = 1
	maxLen         = 39
	illegalPrefix  = "-"
	illegalSuffix  = "-"
	illegalPattern = "--"
)

var legalPattern = regexp.MustCompile("^[-0-9A-Za-z]*$")

func (*GitHub) String() string {
	return "GitHub"
}

func (*GitHub) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		isShortEnough(username) &&
		containsNoIllegalPattern(username) &&
		containsOnlyLegalChars(username) &&
		containsNoIllegalPrefix(username) &&
		containsNoIllegalSuffix(username)
}

func (gh *GitHub) IsAvailable(ctx context.Context, username string) (bool, error) {
	// Test new error: endpoint := fmt.Sprintf("https://githubnkdnakndakadnkadsndsk.com/%s", url.PathEscape(username))
	endpoint := fmt.Sprintf("https://github.com/%s", url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: gh.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := gh.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: gh.String(),
			Cause:    err,
		}
		return false, &err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	return resp.StatusCode == http.StatusNotFound, nil
}

func isShortEnough(username string) bool {
	return utf8.RuneCountInString(username) <= maxLen
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
	return !strings.HasSuffix(username, illegalSuffix)
}
