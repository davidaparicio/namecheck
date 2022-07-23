package namecheck

import (
	"context"
	"fmt"
	"net/http"
)

type Validator interface {
	IsValid(username string) bool
}

type Availabler interface {
	IsAvailable(ctx context.Context, username string) (bool, error)
}

type Checker interface {
	Validator
	Availabler
	fmt.Stringer
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type UnknownAvailabilityError struct {
	Username string
	Platform string
	Cause    error
}

func (e *UnknownAvailabilityError) Error() string {
	const tmpl = "unknown avaibility of %q on %s: %v" //"Please retry after some time or submit an issue on Github"
	return fmt.Sprintf(tmpl, e.Username, e.Platform, e.Cause)
}

func (e *UnknownAvailabilityError) Unwrap() error {
	return e.Cause
}
