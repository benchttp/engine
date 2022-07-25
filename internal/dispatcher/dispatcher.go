package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

var ErrInvalidValue = errors.New("invalid value")

type Dispatcher interface {
	Do(ctx context.Context, maxIter int, callback func()) error
}

type dispatcher struct {
	numWorker int
	sem       *semaphore.Weighted
}

// New returns a Dispatcher initialized with numWorker.
func New(numWorker int) Dispatcher {
	if numWorker < 1 {
		panic(fmt.Sprintf("invalid numWorker value: must be > 1, got %d", numWorker))
	}
	sem := semaphore.NewWeighted(int64(numWorker))
	return dispatcher{sem: sem, numWorker: numWorker}
}

// Do concurrently executes callback at most maxIter times or until ctx is done
// or canceled. Concurrency is handled leveraging the semaphore pattern, which
// ensures at most Dispatcher.numWorkers goroutines are spawned at the same time.
// It returns an early ErrInvalidValue if any of the following conditions is met:
// 	maxIter < 1 and maxIter != -1
// 	maxIter > numWorker
// 	callback == nil
// Else it returns the context error if any or nil.
func (d dispatcher) Do(ctx context.Context, maxIter int, callback func()) error {
	if err := d.validate(maxIter, callback); err != nil {
		return err
	}

	var (
		err error
		wg  sync.WaitGroup
	)

	for i := 0; i < maxIter || maxIter == -1; i++ {
		wg.Add(1)

		if err = d.sem.Acquire(ctx, 1); err != nil {
			// err is either context.DeadlineExceeded or context.Canceled
			// which are expected values so we stop the process silently.
			wg.Done()
			break
		}

		go func() {
			defer func() {
				d.sem.Release(1)
				wg.Done()
			}()
			callback()
		}()
	}

	wg.Wait()
	return err
}

func (d dispatcher) validate(maxIter int, callback func()) error {
	if maxIter < 1 && maxIter != -1 {
		return fmt.Errorf("%w: maxIter: must be -1 or >= 1, got %d", ErrInvalidValue, maxIter)
	}
	if maxIter < d.numWorker && maxIter != -1 {
		return fmt.Errorf(
			"%w: maxIter: must be >= numWorker (%d), got %d",
			ErrInvalidValue, d.numWorker, maxIter,
		)
	}
	if callback == nil {
		return fmt.Errorf("%w: callback: must be non-nil", ErrInvalidValue)
	}
	return nil
}
