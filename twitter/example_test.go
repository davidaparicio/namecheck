package twitter_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/davidaparicio/namecheck/twitter"
)

func ExampleTwitter_IsValid() {
	var tw twitter.Twitter
	fmt.Println(tw.IsValid("eczxaw"))
	// Output: true
}

func ExampleTwitter_IsAvailable() {
	t := twitter.Twitter{
		Client: http.DefaultClient,
	}
	fmt.Println(t.IsAvailable(context.Background(), "dadideo"))
	fmt.Println(t.IsAvailable(context.Background(), "eczxaw"))
	// Output:
	// false <nil>
	// true <nil>
}

// More information https://go.dev/blog/examples | https://go.dev/blog/godoc
