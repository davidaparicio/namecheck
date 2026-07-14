package bluesky_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/bluesky"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*bluesky.Bluesky)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"jay", true},
		{"david-aparicio", true},
		{"ab", false},                         // too short
		{"longer-than-eighteen-chars", false}, // too long
		{"-david", false},                     // illegal prefix
		{"david-", false},                     // illegal suffix
		{"david_aparicio", false},             // illegal char
	}
	var bs bluesky.Bluesky
	for _, c := range cases {
		if got := bs.IsValid(c.username); got != c.want {
			t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
		}
	}
}

func TestIsAvailableUnresolvableHandle(t *testing.T) {
	bs := bluesky.Bluesky{
		Client: stub.ClientWithStatusCodeAndBody(
			http.StatusBadRequest,
			`{"error":"InvalidRequest","message":"Unable to resolve handle"}`,
		),
	}
	avail, err := bs.IsAvailable(context.Background(), "whatever")
	if !avail || err != nil {
		t.Error("IsAvailable must return true on a 400 status code")
	}
}

func TestIsAvailableTaken(t *testing.T) {
	bs := bluesky.Bluesky{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, `{"did":"did:plc:abc"}`),
	}
	avail, err := bs.IsAvailable(context.Background(), "whatever")
	if avail || err != nil {
		t.Error("IsAvailable must return false on a 200 status code")
	}
}

func TestIsAvailableUnexpectedStatusCode(t *testing.T) {
	bs := bluesky.Bluesky{
		Client: stub.ClientWithStatusCode(http.StatusInternalServerError),
	}
	_, err := bs.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on an unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	bs := bluesky.Bluesky{
		Client: stub.ClientWithError(errors.New("oh no")),
	}
	_, err := bs.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on a client error")
	}
}
