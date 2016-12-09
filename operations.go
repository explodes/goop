package operations

import (
	"context"
	"errors"
)

var (
	ErrCancel = errors.New("Cancelled operation")
)

type priority uint

type Op func(ctx context.Context) (interface{}, error)

type opResult struct {
	priority priority
	result   interface{}
	err      error
}
