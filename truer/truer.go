// Package truer provides always True
package truer

import (
	"context"
)

type Truer struct{}

func (*Truer) IsValid(username string) bool {
	return true
}

func (*Truer) IsAvailable(ctx context.Context, username string) (bool, error) {
	return true, nil
}

func (*Truer) String() string {
	return "Truer"
}
