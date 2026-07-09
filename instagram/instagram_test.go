package instagram_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/davidaparicio/namecheck"
	"github.com/davidaparicio/namecheck/instagram"
	"github.com/davidaparicio/namecheck/stub"
)

var _ namecheck.Checker = (*instagram.Instagram)(nil)

func TestIsValid(t *testing.T) {
	cases := []struct {
		desc     string
		username string
		want     bool
	}{
		{"too long", "obviously-longer-than-30-chars-skjdhsdkhfkshkfshdkjfhksdjhf", false},
		{"too short", "ab", false},
		{"illegal chars", "dadi-deo", false},
		{"valid", "dadideo", true},
		{"valid with period and underscore", "da.di_deo", true},
	}
	insta := instagram.Instagram{}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if got := insta.IsValid(c.username); got != c.want {
				t.Errorf("IsValid(%q) = %t; want %t", c.username, got, c.want)
			}
		})
	}
}

func TestUsernameAvailable(t *testing.T) {
	insta := instagram.Instagram{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, "<html><body>Sorry, this page isn't available.</body></html>"),
	}
	avail, err := insta.IsAvailable(context.Background(), "dadideodadideo")
	if err != nil {
		t.Fatalf("IsAvailable: unexpected error: %v", err)
	}
	if !avail {
		t.Error("IsAvailable = false; want true")
	}
}

func TestUsernameNotAvailable(t *testing.T) {
	insta := instagram.Instagram{
		Client: stub.ClientWithStatusCodeAndBody(http.StatusOK, `<meta content="noarchive, noimageindex" name="robots" />`),
	}
	avail, err := insta.IsAvailable(context.Background(), "dadideo")
	if err != nil {
		t.Fatalf("IsAvailable: unexpected error: %v", err)
	}
	if avail {
		t.Error("IsAvailable = true; want false")
	}
}

func TestIsAvailableNot200(t *testing.T) {
	insta := instagram.Instagram{
		Client: stub.ClientWithStatusCode(http.StatusTooManyRequests),
	}
	_, err := insta.IsAvailable(context.Background(), "whatever")
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
	insta := instagram.Instagram{
		Client: stub.ClientWithError(boom),
	}
	_, err := insta.IsAvailable(context.Background(), "whatever")
	var uae *namecheck.UnknownAvailabilityError
	if !errors.As(err, &uae) {
		t.Fatalf("IsAvailable error = %v; want an *UnknownAvailabilityError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("IsAvailable error = %v; want it to wrap %v", err, boom)
	}
}
