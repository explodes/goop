/*
goop is short for Go Operations

goop is a package for executing a series of operations in implicit priority
such that only the highest priority operation is considered a valid result.

There is a case when no operations succeed and the highest priority error is returned
*/
package goop

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrCancel is a special-case error an operation should return when it has been cancelled
	ErrCancel = errors.New("Cancelled operation")
)

// priority is the priority of an Op. Lower is higher.
type priority uint

// opResult is the result of calling an Op, saving its result and call priority
type opResult struct {
	priority priority
	result   interface{}
	err      error
}

// Op is a function that executes in the given context and returns its results
type Op func(ctx context.Context) (interface{}, error)

// execute executes an op and possibly deliver the results to the supplied channel
// as long as the operation does not get trumped
func (op Op) execute(p priority, trumps *naiiveBroadcast, results chan<- *opResult, wg *sync.WaitGroup) {
	defer wg.Done()

	// context to cancel if our results get trumped so that
	// the operation can exit early
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// collect the operations results from another goroutine
	opChan := make(chan *opResult)
	go func() {
		defer close(opChan)
		result, err := op(ctx)

		select {
		case <-ctx.Done():
			return
		default:
			result := &opResult{
				priority: p,
				result:   result,
				err:      err,
			}
			opChan <- result
		}
	}()

	// whichever comes first: higher priority operation
	// finished or we have our results
	for {
		select {
		case n := <-trumps.B():
			// if a higher priority item finished first,
			// ignore the results of our actions
			if n < p {
				return
			}
		case opResult := <-opChan:
			if opResult.err == nil {
				trumps.Broadcast(p)
			}
			results <- opResult
			return
		}
	}
}

// PerformOperations executes a series of functions in implicit priority (the first function is considered highest
// priority, and the last function is considered to have the least priority. Operations with a lower priority will
// be cancelled by Context when a higher priority operation succeeds.
//
// The highest-priority result is returned, or the highest-priority error is returned.
func PerformOperations(ops ...Op) (interface{}, error) {
	if len(ops) == 0 {
		return nil, errors.New("no operations")
	}

	trumps := newBroadcast(len(ops)) // todo: implement a better (real) broadcast system
	defer trumps.Close()

	results := make(chan *opResult)

	wg := &sync.WaitGroup{}
	wg.Add(len(ops))

	var best *opResult

	// Launch our operations
	for p, op := range ops {
		go func(p int, op Op) {
			op.execute(priority(p), trumps, results, wg)
		}(p, op)
	}

	// Collect results
	go func() {
		defer close(results)
		for next := range results {
			if next == nil {
				break
			}
			// Determine if we have a new "best"
			//
			switch {
			case best == nil:
				// Save the first result
				best = next
			case best.result == nil && next.result != nil:
				// We finally have a result, use this one
				best = next
			case best.result != nil && next.result != nil && next.priority < best.priority:
				// Higher priority result received
				best = next
			case best.err != nil && next.err != nil && next.priority < best.priority:
				// Higher priority error received
				best = next
			}
		}
	}()

	wg.Wait()
	results <- nil

	return best.result, best.err
}
