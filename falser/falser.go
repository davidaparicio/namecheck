// Package falser provides always False
package falser

import (
	"context"
)

type Falser struct{}

func (*Falser) IsValid(username string) bool {
	return false
}

func (*Falser) IsAvailable(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (*Falser) String() string {
	return "Falser"
}
