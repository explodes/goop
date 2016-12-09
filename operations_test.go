package goop

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
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

func TestPerformOperations(t *testing.T) {
	if result, err := PerformOperations(); err.Error() != "no operations" {
		t.Fatalf("Unexpected return: (%v, %v)", result, err)
	}
	result, err := PerformOperations(makeOperations(10)...)
	if result.(int) != 5 {
		t.Fatalf("Unexpected response: %d", result)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func BenchmarkPerformOperations(b *testing.B) {
	numOpses := [...]int{1, 10, 100, 1000}
	for _, numOps := range numOpses {
		b.Run(fmt.Sprintf("BenchmarkPerformOperations%d", numOps), func(b *testing.B) {
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
