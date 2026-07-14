package gitlab_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/gitlab"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*gitlab.GitLab)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		username string
		want     bool
	}{
		{"davidaparicio", true},
		{"david.aparicio", true},
		{"david_aparicio", true},
		{"a", false},               // too short
		{"-david", false},          // illegal prefix
		{"david-", false},          // illegal suffix
		{"david.", false},          // illegal suffix
		{"david.git", false},       // illegal suffix
		{"david.atom", false},      // illegal suffix
		{"david--aparicio", false}, // illegal pattern
		{"david aparicio", false},  // illegal char
	}
	var gl gitlab.GitLab
	for _, c := range cases {
		if got := gl.IsValid(c.username); got != c.want {
			t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
		}
	}
}

func TestIsAvailableEmptyResult(t *testing.T) {
	gl := gitlab.GitLab{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, "[]"),
	}
	avail, err := gl.IsAvailable(context.Background(), "whatever")
	if !avail || err != nil {
		t.Error("IsAvailable must return true for an empty result set")
	}
}

func TestIsAvailableNonEmptyResult(t *testing.T) {
	gl := gitlab.GitLab{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, `[{"id":42,"username":"whatever"}]`),
	}
	avail, err := gl.IsAvailable(context.Background(), "whatever")
	if avail || err != nil {
		t.Error("IsAvailable must return false for a non-empty result set")
	}
}

func TestIsAvailableUnexpectedStatusCode(t *testing.T) {
	gl := gitlab.GitLab{
		Client: stub.ClientWithStatusCode(http.StatusTooManyRequests),
	}
	_, err := gl.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on an unexpected status code")
	}
}

func TestIsAvailableClientError(t *testing.T) {
	gl := gitlab.GitLab{
		Client: stub.ClientWithError(errors.New("oh no")),
	}
	_, err := gl.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Error("IsAvailable must return an UnknownAvailabilityError on a client error")
	}
}
