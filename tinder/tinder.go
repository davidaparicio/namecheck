// Package tinder provides primitives to check if an username
// is available on Tinder.
package tinder

import (
	"bytes"
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

type Tinder struct {
	Client namecheck.Client
}

const (
	minLen = 3
	maxLen = 30
)

var legalPattern = regexp.MustCompile("^[-0-9A-Za-z]*$")

func (*Tinder) String() string {
	return "Tinder"
}

func (*Tinder) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen)
}

func (tin *Tinder) IsAvailable(ctx context.Context, username string) (bool, error) {
	endpoint := fmt.Sprintf("https://tinder.com/@marta3%s", url.PathEscape(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: tin.String(),
			Cause:    err,
		}
		return false, &err
	}
	resp, err := tin.Client.Do(req)
	if err != nil {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: tin.String(),
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
			Platform: tin.String(),
			Cause:    err,
		}
		return false, &err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	// Safely read the response body
	bodyBytes, err := io.ReadAll(tee)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %v", err)
	}
	bodyString := string(bodyBytes)

	/* if Log in to like me  = pseudo NOT AVAILABLE (True)
	   else = AVAILABLE (False) */
	return !strings.Contains(bodyString, "</path></svg>Log in to like me</div>"), nil
}
