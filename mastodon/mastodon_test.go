package mastodon_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/mastodon"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*mastodon.Mastodon)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"Gargron", true},
		{"david_aparicio", true},
		{"", false}, // too short
		{"obviously-longer-than-thirty-characters", false}, // too long
		{"david-aparicio", false},                          // illegal char
	}
	var mast mastodon.Mastodon
	for _, c := range cases {
		if got := mast.IsValid(c.username); got != c.want {
			t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
		}
	}
}

func TestIsAvailableNotFound(t *testing.T) {
	mast := mastodon.Mastodon{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusNotFound, `{"error":"Record not found"}`),
	}
	avail, err := mast.IsAvailable(context.Background(), "whatever")
	if !avail || err != nil {
		t.Error("IsAvailable must return true on a 404 status code")
	}
}

func TestIsAvailableTaken(t *testing.T) {
	mast := mastodon.Mastodon{
		Client: stub.ClientWithStatusCode(http.StatusOK),
	}
	avail, err := mast.IsAvailable(context.Background(), "whatever")
	if avail || err != nil {
		t.Error("IsAvailable must return false on a 200 status code")
	}
}

func TestIsAvailableUnexpectedStatusCode(t *testing.T) {
	mast := mastodon.Mastodon{
		Client: stub.ClientWithStatusCode(http.StatusTooManyRequests),
	}
	_, err := mast.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on an unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	mast := mastodon.Mastodon{
		Client: stub.ClientWithError(errors.New("oh no")),
	}
	_, err := mast.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on a client error")
	}
}
