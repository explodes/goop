package main

// +build sample

import (
	"context"
	"errors"
	"fmt"
	"github.com/explodes/operations"
	"math/rand"
	"time"
)

func sampleOperation(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, operations.ErrCancel
	case <-time.After(time.Duration(10+rand.Intn(10)) * time.Millisecond):
		if rand.Float32() > 0.90 {
			return nil, errors.New("Random failure")
		}
		return rand.Intn(1000), nil
	}
}

func main() {
	rand.Seed(1001)

	const N = 300

	ops := make([]operations.Op, N)
	for i := 0; i < N; i++ {
		ops[i] = sampleOperation
	}

	result, err := operations.PerformOperations(ops...)
	fmt.Println(result, err)
}
