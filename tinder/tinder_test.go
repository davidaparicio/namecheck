package tinder_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/stub"
	"github.com/davidaparicio/namecheck/tinder"
)

var _ namecheck.Checker = (*tinder.Tinder)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		desc     string
		username string
		want     bool
	}{
		{"too long", "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf", false},
		{"too short", "ab", false},
		{"illegal chars", "marta_1789", false},
		{"valid", "marta1789", true},
		{"valid with hyphen", "marta-1789", true},
	}
	tin := tinder.Tinder{}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if got := tin.IsValid(c.username); got != c.want {
				t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
			}
		})
	}
}

func TestUsernameAvailable(t *testing.T) {
	tin := tinder.Tinder{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, "<html><body>Looking for someone?</body></html>"),
	}
	avail, err := tin.IsAvailable(context.Background(), "marta1789")
	if err != nil {
		t.Fatalf("IsAvailable: unexpected error: %v", err)
	}
	if !avail {
		t.Error("IsAvailable = false; want true")
	}
}

func TestUsernameNotAvailable(t *testing.T) {
	tin := tinder.Tinder{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, `<div><svg><path></path></svg>Log in to like me</div>`),
	}
	avail, err := tin.IsAvailable(context.Background(), "marta")
	if err != nil {
		t.Fatalf("IsAvailable: unexpected error: %v", err)
	}
	if avail {
		t.Error("IsAvailable = true; want false")
	}
}

func TestIsAvailableNot200(t *testing.T) {
	tin := tinder.Tinder{
		Client: stub.ClientWithStatusCode(http.StatusNotFound),
	}
	_, err := tin.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Fatalf("IsAvailable error = %v; want an *UnknownAvailabilityError", err)
	}
	if uae.Cause == nil {
		t.Error("UnknownAvailabilityError.Cause is nil; want the unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	boom := errors.New("network down")
	tin := tinder.Tinder{
		Client: stub.ClientWithError(boom),
	}
	_, err := tin.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Fatalf("IsAvailable error = %v; want an *UnknownAvailabilityError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("IsAvailable error = %v; want it to wrap %v", err, boom)
	}
}
