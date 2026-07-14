package hackernews_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/hackernews"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*hackernews.HackerNews)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"pg", true},
		{"david_aparicio", true},
		{"a", false},                    // too short
		{"longer-than-15-chars", false}, // too long
		{"david aparicio", false},       // illegal char
	}
	var hn hackernews.HackerNews
	for _, c := range cases {
		if got := hn.IsValid(c.username); got != c.want {
			t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
		}
	}
}

func TestIsAvailableNullBody(t *testing.T) {
	hn := hackernews.HackerNews{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, "null"),
	}
	avail, err := hn.IsAvailable(context.Background(), "whatever")
	if !avail || err != nil {
		t.Error("IsAvailable must return true on a null body")
	}
}

func TestIsAvailableTaken(t *testing.T) {
	hn := hackernews.HackerNews{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, `{"id":"whatever","karma":1}`),
	}
	avail, err := hn.IsAvailable(context.Background(), "whatever")
	if avail || err != nil {
		t.Error("IsAvailable must return false on a non-null body")
	}
}

func TestIsAvailableUnexpectedStatusCode(t *testing.T) {
	hn := hackernews.HackerNews{
		Client: stub.ClientWithStatusCode(http.StatusInternalServerError),
	}
	_, err := hn.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on an unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	hn := hackernews.HackerNews{
		Client: stub.ClientWithError(errors.New("oh no")),
	}
	_, err := hn.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on a client error")
	}
}
