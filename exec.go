package operations

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
)

func executeOperation(p priority, op Op, done Broadcast, result chan<- *opResult, wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("executeOperation", p, "done")
		wg.Done()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opChan := make(chan *opResult)

	// run the operation in a goroutine
	go func() {
		defer close(opChan)

		result, err := op(ctx)
		select {
		case <-ctx.Done():
			fmt.Println("executeOperation go1", p, "not sending results: Done")
			return
		default:
			result := &opResult{
				priority: p,
				result:   result,
				err:      err,
			}
			opChan <- result
			fmt.Println("executeOperation go1", p, "sending ", result)
		}
	}()

	// whichever comes first:
	//  - higher priority operation finished
	//  - we have our results
	for {
		select {
		case n := <-done.B():
			// a higher priority item finished first, ignore the results of our actions
			if n < p {
				fmt.Println("executeOperation", p, "trumped by", n)
				result <- nil
				return
			}
		case opResult := <-opChan:
			if opResult.err == nil {
				fmt.Println("executeOperation", p, "broadcast", p)
				done.Broadcast(p)
			}
			fmt.Println("executeOperation", p, "send", opResult)
			result <- opResult
			return
		}
	}
}

func PerformOperations(ops ...Op) (interface{}, error) {

	doneBroadcast := newBroadcast(len(ops))
	defer doneBroadcast.Close()

	resultsChan := make(chan *opResult)

	wg := &sync.WaitGroup{}
	wg.Add(len(ops))

	resultsHeap := make(opResults, 0)

	for p, op := range ops {
		fmt.Println("PerformOperations", "executeOperation", p)
		go func(p int, op Op) {
			executeOperation(priority(p), op, doneBroadcast, resultsChan, wg)
		}(p, op)
	}

	// collect and heapify results
	go func() {
		defer close(resultsChan)
		for result := range resultsChan {
			if result == nil {
				continue
			}
			fmt.Println("PerformOperations", "collect", result)
			heap.Push(&resultsHeap, result)
		}
	}()

	fmt.Println("PerformOperations", "waiting")
	wg.Wait()
	resultsChan <- nil
	fmt.Println("PerformOperations", "waited")

	for _, result := range resultsHeap {
		fmt.Println("PerformOperations", "dump result", *result)
	}

	// we have our results, return the first result or the first error
	for _, result := range resultsHeap {
		fmt.Println("PerformOperations", "check result", *result)
		if result.err == nil {
			return result.result, nil
		}
	}

	// no successful operations, return the first error
	firstResult := resultsHeap[0]
	if firstResult.err == ErrCancel {
		return nil, nil
	}
	return nil, firstResult.err
}
