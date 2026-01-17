package work

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

// Job represents a unit of work to be processed by a queue worker
type Job any

// Worker is a function that processes a job
type Worker[T Job] func(context.Context, T)

// Options represents the configuration options for a Queue
type Options struct {
	// Size is the size of the job queue buffer
	// defaults to Workers * 4
	Size int
	// Workers is the number of concurrent workers to process jobs
	// defaults to number of CPU cores
	Workers int
}

// queue is the internal queue
type queue[T Job] struct {
	closed   atomic.Bool
	nWorkers int
	queue    chan T
	sem      chan struct{}
}

// newQueue creates a new internal queue with the given options
func newQueue[T Job](options ...Options) *queue[T] {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	}
	if opts.Size <= 0 {
		opts.Size = opts.Workers * 4
	}
	return &queue[T]{
		nWorkers: opts.Workers,
		queue:    make(chan T, opts.Size),
		sem:      make(chan struct{}, opts.Workers),
	}
}

// close closes the internal queue channels
func (q *queue[T]) close() bool {
	if q.closed.CompareAndSwap(false, true) {
		close(q.queue)
		close(q.sem)
		return true
	}
	return false
}

// Push adds a job to the queue
// returns false if the queue is full and the job was not added
func (q *queue[T]) Push(job T) bool {
	select {
	case q.queue <- job:
		return true
	default:
		return false
	}
}

// Queue represents a work queue that processes jobs using a worker function
type Queue[T Job] struct {
	*queue[T]
	worker Worker[T]
}

// NewQueue creates a new Queue with the given worker and options
func NewQueue[T Job](worker Worker[T], options ...Options) *Queue[T] {
	return &Queue[T]{
		queue:  newQueue[T](options...),
		worker: worker,
	}
}

// Run starts the queue and begins processing jobs
// runs until the context is cancelled
func (q *Queue[T]) Run(ctx context.Context) error {
	if q.closed.Load() {
		return errors.New("queue is closed")
	}
	if q.worker == nil {
		return errors.New("worker must be provided")
	}

	wg := sync.WaitGroup{}
	for range q.nWorkers {
		wg.Go(func() {
			q.runWorker(ctx)
		})
	}

	<-ctx.Done()
	wg.Wait() // wait for all workers to finish
	q.close()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		return err
	}
	return nil
}

// runWorker processes jobs from the queue until the context is cancelled
func (q *Queue[T]) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-q.queue.queue:
			q.sem <- struct{}{} // acquire
			func() {
				defer func() { <-q.sem }() // release even on panic
				// process job
				q.worker(ctx, job)
			}()
		}
	}
}

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
