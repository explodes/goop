package goop

import (
	"context"
	"errors"
	"testing"
	"time"
	"fmt"
)

func makeOperation(p int, N int) Op {
	return func(ctx context.Context) (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ErrCancel
		case <-time.After(10 * time.Millisecond):
			if p < N/2 {
				return nil, errors.New("Doomed to fail")
			}
			return p, nil
		}
	}
}

func makeOperations(n int) []Op {
	ops := make([]Op, n)
	for p := 0; p < n; p++ {
		ops[p] = makeOperation(p, n)
	}
	return ops
}

func BenchmarkPerformOperations(b *testing.B) {
	numOpses := [...]int{1, 10, 100, 1000}
	for _, numOps := range numOpses {
		b.Run(fmt.Sprintf("BenchmarkPerformOperations%d",numOps), func(b *testing.B) {
			benchmarkOperations(numOps, b)
		})
	}
}

func benchmarkOperations(n int, b *testing.B) {
	for i := 0; i < b.N; i++ {
		ops := makeOperations(n)
		PerformOperations(ops...)
	}
}
