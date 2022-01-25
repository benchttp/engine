package semimpl

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Do concurrently executes callback at most maxIter times or until ctx is done
// or canceled. Concurrency is handled leveraging the semaphore pattern, which
// ensures at most numWorkers goroutines are spawned at the same time.
func Do(ctx context.Context, numWorkers, maxIter int, callback func()) {
	numWorkers = sanitizeNumWorkers(numWorkers)
	maxIter = sanitizeMaxIter(maxIter)
	callback = sanitizeCallback(callback)

	sem := semaphore.NewWeighted(int64(numWorkers))
	wg := sync.WaitGroup{}

	for i := 0; i < maxIter || maxIter == 0; i++ {
		wg.Add(1)

		if err := sem.Acquire(ctx, 1); err != nil {
			// err is either context.DeadlineExceeded or context.Canceled
			// which are expected values so we stop the process silently.
			wg.Done()
			break
		}

		go func() {
			defer func() {
				sem.Release(1)
				wg.Done()
			}()
			callback()
		}()
	}

	wg.Wait()
}

func sanitizeNumWorkers(numWorkers int) int {
	if numWorkers < 1 {
		return 1
	}
	return numWorkers
}

func sanitizeMaxIter(maxIter int) int {
	if maxIter < 0 {
		return 0
	}
	return maxIter
}

func sanitizeCallback(callback func()) func() {
	if callback == nil {
		return func() {}
	}
	return callback
}
