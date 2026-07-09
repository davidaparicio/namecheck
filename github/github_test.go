package github_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/github"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*github.GitHub)(nil)

func TestUsernameTooLong(t *testing.T) {
	var gh = github.GitHub{}
	username := "obviously-longer-than-39-chars-skjdhsdkhfkshkfshdkjfhksdjhf"
	want := false
	got := gh.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestIsAvailableNot200(t *testing.T) {
	gh := github.GitHub{
		Client: stub.ClientWithStatusCode(http.StatusOK),
	}
	username := "whatever"
	avail, err := gh.IsAvailable(context.Background(), username)
	if avail || err != nil {
		t.Error("IsAvailable must return a 404 status code")
	}
}

func TestIsAvailable404(t *testing.T) {
	gh := github.GitHub{
		Client: stub.ClientWithStatusCode(http.StatusNotFound),
	}
	avail, err := gh.IsAvailable(context.Background(), "whatever")
	if err != nil {
		t.Fatalf("IsAvailable: unexpected error: %v", err)
	}
	if !avail {
		t.Error("IsAvailable = false; want true when the profile page is a 404")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	boom := errors.New("network down")
	gh := github.GitHub{
		Client: stub.ClientWithError(boom),
	}
	_, err := gh.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Fatalf("IsAvailable error = %v; want an *UnknownAvailabilityError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("IsAvailable error = %v; want it to wrap %v", err, boom)
	}
}
