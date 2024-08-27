package falser_test

import (
	"context"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/falser"
)

var _ namecheck.Checker = (*falser.Falser)(nil)

func TestUsernameTooLong(t *testing.T) {
	falser := falser.Falser{}
	username := "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf"
	want := false
	got := falser.IsValid(username)
	if got != want {
		t.Errorf(
			"IsValid(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}

func TestUsernameNotAvailable(t *testing.T) {
	falser := falser.Falser{}
	username := "test"
	want := false
	got, err := falser.IsAvailable(context.Background(), username)
	if got != want && err == nil {
		t.Errorf(
			"IsAvailable(%s) = %t; want %t",
			username,
			got,
			want,
		)
	}
}
