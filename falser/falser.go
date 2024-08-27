// Package falser provides always False
package falser

import (
	"context"

	"github.com/davidaparicio/namecheck"
)

type Falser struct {
	Client namecheck.Client
}

func (*Falser) IsValid(username string) bool {
	return false
}

func (*Falser) IsAvailable(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (*Falser) String() string {
	return "Falser"
}
