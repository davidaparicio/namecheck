package tinder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/tinder"
)

var _ namecheck.Checker = (*tinder.Tinder)(nil)

func TestUsernameTooLong(t *testing.T) {
	tin := tinder.Tinder{}
	username := "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf"
	want := false
	got := tin.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameAvailable(t *testing.T) {
	tin := tinder.Tinder{Client: http.DefaultClient}
	username := "marta1789"
	want := true
	got, err := tin.IsAvailable(context.Background(), username)
	if got != want && err == nil {
		t.Errorf(
			"IsAvailable(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameNotAvailable(t *testing.T) {
	tin := tinder.Tinder{Client: http.DefaultClient}
	username := "marta"
	want := false
	got, err := tin.IsAvailable(context.Background(), username)
	if got != want && err == nil {
		t.Errorf(
			"IsAvailable(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

/*func TestIsAvailableNot200(t *testing.T) {
	insta := instagram.Instagram{
		Client: stub.ClientWithStatusCode(http.StatusNotFound),
	}
	username := "whatever"
	avail, err := insta.IsAvailable(context.Background(), username)
	if avail || err != nil {
		t.Error("IsAvailable must return a 404 status code")
	}
}*/
