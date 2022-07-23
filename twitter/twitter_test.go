package twitter_test

import (
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/twitter"
)

var _ namecheck.Checker = (*twitter.Twitter)(nil)

func TestUsernameTooLong(t *testing.T) {
	tw := twitter.Twitter{}
	username := "obviously_longer_than_15_chars"
	want := false
	got := tw.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameTooShort(t *testing.T) {
	tw := twitter.Twitter{}
	username := "foo"
	want := false
	got := tw.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameContainsIllegalPattern(t *testing.T) {
	tw := twitter.Twitter{}
	username := "FtWittEroo"
	want := false
	got := tw.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameContainsIllegalChars(t *testing.T) {
	tw := twitter.Twitter{}
	username := "jub0bs-"
	want := false
	got := tw.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameValid(t *testing.T) {
	tw := twitter.Twitter{}
	username := "jub0bs"
	want := true
	got := tw.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}
