package work

import (
	"context"
	"errors"
	"sync"
)

// Job represents a unit of work to be processed by a queue worker
type Job any

// Worker is a function that processes a job
type Worker[T Job] func(context.Context, T) error

// JobQueue is a queue that processes jobs using a worker function
type JobQueue[T Job] struct {
	queue   chan T
	worker  Worker[T]
	workers int
}

// Queue is an interface for pushing jobs to a queue
type Queue[T Job] interface {
	// Push adds a job to the queue, returning false if the queue is full
	Push(T) bool
}

// NewJobQueue creates a new JobQueue with the specified number of workers,
// queue buffer size and worker function
// if workers is 0 or negative, it defaults to the number of CPU cores
// if size is 0 or negative, it defaults to workers * 4
func NewJobQueue[T Job](workers int, size int, worker Worker[T]) *JobQueue[T] {
	if workers <= 0 {
		workers = 1
	}
	if size <= 0 {
		size = workers * 4
	}
	return &JobQueue[T]{
		workers: workers,
		queue:   make(chan T, size),
		worker:  worker,
	}
}

// Close closes the job queue, preventing any new jobs from being added
// after calling Close, the queue will panic if Push is called
func (q *JobQueue[T]) Close() {
	close(q.queue)
}

// Push adds a job to the queue
// returns false if the queue is full and the job cannot be added
func (q *JobQueue[T]) Push(job T) bool {
	select {
	case q.queue <- job:
		return true
	default:
		return false
	}
}

// Run starts processing jobs from the queue using the worker function
// it blocks until the context is canceled or an error occurs in a worker
func (q *JobQueue[T]) Run(ctx context.Context) error {
	if q.worker == nil {
		return errors.New("worker must be provided")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, 1)

	var wg sync.WaitGroup

	for i := 0; i < q.workers; i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return

				case job, ok := <-q.queue:
					if !ok {
						return // queue closed, exit worker
					}

					if err := q.worker(ctx, job); err != nil {
						select {
						case errs <- err: // emit error
							cancel()
						default:
						}
						return
					}
				}
			}
		})
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errs:
		<-done
		return err

	case <-done:
		return nil

	case <-ctx.Done():
		<-done

		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}
}
