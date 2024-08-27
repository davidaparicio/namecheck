// Package truer provides always True
package truer

import (
	"context"

	"github.com/davidaparicio/namecheck"
)

type Truer struct {
	Client namecheck.Client
}

func (*Truer) IsValid(username string) bool {
	return true
}

func (*Truer) IsAvailable(ctx context.Context, username string) (bool, error) {
	return true, nil
}

func (*Truer) String() string {
	return "Truer"
}
