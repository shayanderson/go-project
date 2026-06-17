package work

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	// ErrQueueAlreadyRunning is returned when trying to run a queue that is already running
	ErrQueueAlreadyRunning = errors.New("queue is already running")
	// ErrQueueWorkerRequired is returned when trying to run a queue without a worker
	ErrQueueWorkerRequired = errors.New("worker must be provided")
)

// Job represents a unit of work to be processed by a queue worker
type Job any

// Worker is a function that processes a job
type Worker[T Job] func(context.Context, T) error

// JobQueue is a queue that processes jobs using a worker function
type JobQueue[T Job] struct {
	closed  bool
	mu      sync.RWMutex
	queue   chan T
	running atomic.Bool
	worker  Worker[T]
	workers int
}

// Queue is an interface for pushing jobs to a queue
type Queue[T Job] interface {
	// Push adds a job to the queue, returning false if the queue is full
	Push(T) bool
}

// QueueOptions represents the options for creating a queue
type QueueOptions[T Job] struct {
	// Size is the buffer size of the queue channel
	Size int
	// Worker is the function that processes jobs from the queue
	Worker Worker[T]
	// Workers is the number of worker goroutines to process jobs from the queue
	Workers int
}

// NewJobQueue creates a new JobQueue with the specified number of workers,
// queue buffer size and worker function
// if workers is 0 or negative, it defaults to the number of CPU cores
// if size is 0 or negative, it defaults to workers * 4
func NewJobQueue[T Job](opts QueueOptions[T]) *JobQueue[T] {
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if opts.Size <= 0 {
		opts.Size = opts.Workers * 4
	}
	return &JobQueue[T]{
		workers: opts.Workers,
		queue:   make(chan T, opts.Size),
		worker:  opts.Worker,
	}
}

// Close closes the queue and prevents new jobs from being added
// buffered jobs already in the queue are still processed
// subsequent calls to Push return false
func (q *JobQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	close(q.queue)
}

// Push adds a job to the queue
// returns false if the queue is full or the queue is closed and the job cannot be added
func (q *JobQueue[T]) Push(job T) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return false
	}

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
		return ErrQueueWorkerRequired
	}
	if !q.running.CompareAndSwap(false, true) {
		return ErrQueueAlreadyRunning
	}
	defer q.running.Store(false)

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
