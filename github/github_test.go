package github_test

import (
	"context"
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
