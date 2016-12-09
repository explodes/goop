package main

// +build sample

import (
	"context"
	"errors"
	"fmt"
	"github.com/explodes/goop"
	"math/rand"
	"time"
	"os"
	"runtime/pprof"
)

func makeOperation(p int, N int) goop.Op {
	return func(ctx context.Context) (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, goop.ErrCancel
		case <-time.After(10 * time.Millisecond):
			if p < N / 2 {
				return nil, errors.New("Doomed to fail")
			}
			return p, nil
		}
	}
}

func profile() {
	pprof.Lookup("heap").WriteTo(os.Stdout, 1)
	for {
		select {
		case <-time.After(1000 * time.Millisecond):
		//pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			pprof.Lookup("heap").WriteTo(os.Stdout, 1)
		}
	}
}

func main() {
	rand.Seed(1002)

	//go profile()

	const N = 3000

	ops := make([]goop.Op, N)
	for i := 0; i < N; i++ {
		ops[i] = makeOperation(i, N)
	}

	result, err := goop.PerformOperations(ops...)
	fmt.Println(result, err)
}
