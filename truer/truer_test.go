package truer_test

import (
	"context"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/truer"
)

var _ namecheck.Checker = (*truer.Truer)(nil)

func TestUsernameLongIsOK(t *testing.T) {
	truer := truer.Truer{}
	username := "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf"
	want := true
	got := truer.IsValid(username)
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
	truer := truer.Truer{}
	username := "test"
	want := true
	got, err := truer.IsAvailable(context.Background(), username)
	if got != want && err == nil {
		t.Errorf(
			"IsAvailable(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}
