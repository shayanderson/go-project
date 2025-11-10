package work

import (
	"context"
	"errors"
	"sync"
)

// ErrWorker is a function that processes a job and returns an error if one occurs
type ErrWorker[T Job] func(context.Context, T) error

// ErrQueue is a queue that processes jobs with error handling
type ErrQueue[T Job] struct {
	*queue[T]
	errors chan error
	worker ErrWorker[T]
}

// NewErrQueue creates a new ErrQueue with the given worker and options
func NewErrQueue[T Job](worker ErrWorker[T], options ...Options) *ErrQueue[T] {
	q := newQueue[T](options...)
	return &ErrQueue[T]{
		errors: make(chan error, cap(q.queue)),
		queue:  q,
		worker: worker,
	}
}

// Run starts the ErrQueue and begins processing jobs using the worker function
// runs until the context is cancelled or an error occurs
func (q *ErrQueue[T]) Run(ctx context.Context) error {
	if q.closed.Load() {
		return errors.New("queue is closed")
	}
	if q.worker == nil {
		return errors.New("worker must be provided")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}
	for range q.nWorkers {
		wg.Go(func() {
			q.runWorker(ctx)
		})
	}

	var err error
	select {
	case <-ctx.Done():
	case err = <-q.errors:
		// an error occurred, stop workers
		cancel()
	}
	wg.Wait() // wait for all workers to finish
	if q.close() {
		close(q.errors)
	}

	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil && err != context.Canceled {
		return err
	}
	return nil
}

// runWorker processes jobs from the queue until the context is cancelled or an error occurs
func (q *ErrQueue[T]) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-q.queue.queue:
			q.sem <- struct{}{} // acquire
			func() {
				defer func() { <-q.sem }() // release even on panic
				// process job
				if err := q.worker(ctx, job); err != nil {
					q.errors <- err
					return
				}
			}()
		}
	}
}
