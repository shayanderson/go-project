package work

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var processed atomic.Int32

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    2,
		Worker: func(context.Context, int) error {
			processed.Add(1)
			return nil
		},
	},
	)

	errCh := make(chan error, 1)

	go func() {
		errCh <- q.Run(ctx)
	}()

	if !q.Push(1) {
		t.Fatal("expected push 1 to succeed")
	}

	if !q.Push(2) {
		t.Fatal("expected push 2 to succeed")
	}

	for processed.Load() < 2 {
		time.Sleep(time.Millisecond)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if processed.Load() != 2 {
		t.Fatalf("expected 2 processed jobs, got %d", processed.Load())
	}
}

func TestQueueDefaults(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 0,
		Size:    0,
		Worker: func(context.Context, int) error {
			return nil
		},
	})

	if q.workers != 1 {
		t.Fatalf("expected default workers to be 1, got %d", q.workers)
	}

	if cap(q.queue) != 4 {
		t.Fatalf("expected default queue size to be 4, got %d", cap(q.queue))
	}
}

func TestQueueFull(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(context.Context, int) error {
			time.Sleep(time.Second)
			return nil
		},
	})

	if !q.Push(1) {
		t.Fatal("expected first push to succeed")
	}

	if q.Push(2) {
		t.Fatal("expected second push to fail")
	}
}

func TestQueueWorkerError(t *testing.T) {
	t.Parallel()

	expected := errors.New("test error")

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    2,
		Worker: func(context.Context, int) error {
			return expected
		},
	})
	defer q.Close()

	errs := make(chan error, 1)

	go func() {
		errs <- q.Run(t.Context())
	}()

	time.Sleep(time.Millisecond)

	if !q.Push(1) {
		t.Fatal("expected push to succeed")
	}

	err := <-errs

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestQueueNilWorker(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
	})

	err := q.Run(t.Context())

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrQueueWorkerRequired) {
		t.Fatalf("expected %v, got %v", ErrQueueWorkerRequired, err)
	}
}

func TestQueueClose(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(context.Context, int) error {
			return nil
		},
	})

	q.Close()

	err := q.Run(t.Context())

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	q.Close() // should not panic
}

func TestQueueContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(context.Context, int) error {
			return nil
		},
	})

	errCh := make(chan error, 1)

	go func() {
		errCh <- q.Run(ctx)
	}()

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestQueueDeadlineExceeded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Millisecond,
	)
	defer cancel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(context.Context, int) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	})

	err := q.Run(ctx)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"expected %v, got %v",
			context.DeadlineExceeded,
			err,
		)
	}
}

func TestQueuePushAfterClose(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    10,
		Worker: func(ctx context.Context, i int) error {
			return nil
		},
	})

	done := make(chan any, 1)

	go func() {
		defer func() {
			done <- recover()
		}()

		for {
			if !q.Push(1) {
				return
			}
		}
	}()

	time.Sleep(time.Millisecond)
	q.Close()

	if r := <-done; r != nil {
		t.Fatalf("expected no panic, got %v", r)
	}
}

func TestQueuePushClosed(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(ctx context.Context, i int) error {
			return nil
		},
	})

	q.Close()

	if q.Push(1) {
		t.Fatal("expected push to fail on closed queue")
	}
}

func TestQueueAlreadyRunning(t *testing.T) {
	t.Parallel()

	q := NewJobQueue(QueueOptions[int]{
		Workers: 1,
		Size:    1,
		Worker: func(ctx context.Context, i int) error {
			return nil
		},
	})

	errCh := make(chan error, 1)

	go func() {
		errCh <- q.Run(t.Context())
	}()

	time.Sleep(time.Millisecond)

	err := q.Run(t.Context())

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrQueueAlreadyRunning) {
		t.Fatalf("expected %v, got %v", ErrQueueAlreadyRunning, err)
	}

	q.Close()

	if err := <-errCh; err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestQueueWorkers(t *testing.T) {
	t.Parallel()

	const workers = 4
	const jobs = 8

	var running atomic.Int32
	var maxRunning atomic.Int32
	started := make(chan struct{}, jobs)

	q := NewJobQueue(QueueOptions[int]{
		Workers: workers,
		Size:    jobs,
		Worker: func(ctx context.Context, i int) error {
			n := running.Add(1)
			defer running.Add(-1)

			for {
				old := maxRunning.Load()
				if n <= old || maxRunning.CompareAndSwap(old, n) {
					break
				}
			}

			started <- struct{}{}

			time.Sleep(10 * time.Millisecond)

			return nil
		},
	})

	errCh := make(chan error, 1)

	go func() {
		errCh <- q.Run(t.Context())
	}()

	for i := range jobs {
		if !q.Push(i) {
			t.Fatalf("expected push %d to succeed", i)
		}
	}

	for range jobs {
		<-started
	}

	q.Close()

	if err := <-errCh; err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got := maxRunning.Load(); got != workers {
		t.Fatalf("expected max running workers to be %d, got %d", workers, got)
	}
}
