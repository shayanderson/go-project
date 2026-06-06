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

	q := NewQueue(
		1,
		2,
		func(context.Context, int) error {
			processed.Add(1)
			return nil
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

	q := NewQueue(0, 0, func(context.Context, int) error {
		return nil
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

	q := NewQueue(
		1,
		1,
		func(context.Context, int) error {
			time.Sleep(time.Second)
			return nil
		},
	)

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

	q := NewQueue(
		1,
		2,
		func(context.Context, int) error {
			return expected
		},
	)

	if !q.Push(1) {
		t.Fatal("expected push to succeed")
	}

	q.Close()

	err := q.Run(t.Context())

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestQueueNilWorker(t *testing.T) {
	t.Parallel()

	q := NewQueue[int](1, 1, nil)

	err := q.Run(t.Context())

	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "worker must be provided" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueueClose(t *testing.T) {
	t.Parallel()

	q := NewQueue(
		1,
		1,
		func(context.Context, int) error {
			return nil
		},
	)

	q.Close()

	err := q.Run(t.Context())

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestQueueContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	q := NewQueue(
		1,
		1,
		func(context.Context, int) error {
			return nil
		},
	)

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

	q := NewQueue(
		1,
		1,
		func(context.Context, int) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	)

	err := q.Run(ctx)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"expected %v, got %v",
			context.DeadlineExceeded,
			err,
		)
	}
}
