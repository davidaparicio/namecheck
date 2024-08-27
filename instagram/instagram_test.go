package instagram_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/instagram"
)

var _ namecheck.Checker = (*instagram.Instagram)(nil)

func TestUsernameTooLong(t *testing.T) {
	insta := instagram.Instagram{}
	username := "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf"
	want := false
	got := insta.IsValid(username)
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
	insta := instagram.Instagram{Client: http.DefaultClient}
	username := "dadideodadideo"
	want := true
	got, err := insta.IsAvailable(context.Background(), username)
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
	insta := instagram.Instagram{Client: http.DefaultClient}
	username := "dadideo"
	want := false
	got, err := insta.IsAvailable(context.Background(), username)
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
