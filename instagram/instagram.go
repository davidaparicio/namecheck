// Package instagram provides primitives to check if an username
// is available on Instagram.
package instagram

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

type Instagram struct {
	Client namecheck.Client
}

const (
	minLen = 3
	maxLen = 30
)

var legalPattern = regexp.MustCompile("^[-0-9A-Za-z]*$")

func (*Instagram) String() string {
	return "Instagram"
}

func (*Instagram) IsValid(username string) bool {
	return internal.IsLongEnough(username, minLen) &&
		internal.IsShortEnough(username, maxLen)
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
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err := namecheck.UnknownAvailabilityError{
			Username: username,
			Platform: insta.String(),
			Cause:    err,
		}
		return false, &err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	// Step 4: Safely read the response body
	bodyBytes, err := io.ReadAll(tee)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %v", err)
	}
	bodyString := string(bodyBytes)

	//fmt.Println(bodyString)
	//strings.Contains(bodyString, "Sorry, this page isn't available."))

	/* if !noimageindex  = pseudo AVAILABLE (True)
	   else noimageindex = NOT AVAILABLE (False) */
	return !strings.Contains(bodyString, "noarchive, noimageindex"), nil
}
