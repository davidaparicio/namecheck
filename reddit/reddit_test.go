package reddit_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/reddit"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*reddit.Reddit)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"spez", true},
		{"david-aparicio", true},
		{"david_aparicio", true},
		{"ab", false},                            // too short
		{"longer-than-twenty-characters", false}, // too long
		{"david aparicio", false},                // illegal char
	}
	var red reddit.Reddit
	for _, c := range cases {
		if got := red.IsValid(c.username); got != c.want {
			t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
		}
	}
}

func TestIsAvailableNotFound(t *testing.T) {
	red := reddit.Reddit{
		Client: stub.ClientWithStatusCode(http.StatusNotFound),
	}
	avail, err := red.IsAvailable(context.Background(), "whatever")
	if !avail || err != nil {
		t.Error("IsAvailable must return true on a 404 status code")
	}
}

func TestIsAvailableTaken(t *testing.T) {
	red := reddit.Reddit{
		Client: stub.ClientWithStatusCode(http.StatusOK),
	}
	avail, err := red.IsAvailable(context.Background(), "whatever")
	if avail || err != nil {
		t.Error("IsAvailable must return false on a 200 status code")
	}
}

func TestIsAvailableUnexpectedStatusCode(t *testing.T) {
	red := reddit.Reddit{
		Client: stub.ClientWithStatusCode(http.StatusTooManyRequests),
	}
	_, err := red.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on an unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	red := reddit.Reddit{
		Client: stub.ClientWithError(errors.New("oh no")),
	}
	_, err := red.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on a client error")
	}
}
