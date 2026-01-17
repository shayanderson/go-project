package work

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shayanderson/go-project/v2/internal/test"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var processed atomic.Int32
	q := NewQueue(func(context.Context, int) {
		processed.Add(1)
	}, Options{Workers: 1, Size: 2})

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Run(ctx)
	}()

	pushed := q.Push(1)
	test.True(t, pushed)
	pushed = q.Push(2)
	test.True(t, pushed)

	for processed.Load() < 2 {
		time.Sleep(5 * time.Millisecond)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if processed.Load() != 2 {
		t.Fatalf("expected 2 processed jobs, got %d", processed.Load())
	}
}

func TestQueueRunClosed(t *testing.T) {
	q := NewQueue(func(context.Context, int) {})
	q.close()

	err := q.Run(t.Context())
	if err == nil {
		t.Fatal("expected error when running a closed queue")
	}
	if err.Error() != "queue is closed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var processed atomic.Int32
	q := NewErrQueue(func(context.Context, int) error {
		processed.Add(1)
		return nil
	}, Options{Workers: 1, Size: 2})

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Run(ctx)
	}()

	pushed := q.Push(1)
	test.True(t, pushed)
	pushed = q.Push(2)
	test.True(t, pushed)

	for processed.Load() < 2 {
		time.Sleep(5 * time.Millisecond)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if processed.Load() != 2 {
		t.Fatalf("expected 2 processed jobs, got %d", processed.Load())
	}
}

func TestErrQueueError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var processed atomic.Int32
	q := NewErrQueue(func(context.Context, int) error {
		processed.Add(1)
		if processed.Load() == 2 {
			return errors.New("test error")
		}
		return nil
	}, Options{Workers: 1, Size: 2})

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Run(ctx)
	}()

	q.Push(1)
	q.Push(2)

	for processed.Load() < 2 {
		time.Sleep(5 * time.Millisecond)
	}

	cancel()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "test error" {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed.Load() != 2 {
		t.Fatalf("expected 2 processed jobs, got %d", processed.Load())
	}
}

func TestErrQueueRunClosed(t *testing.T) {
	q := NewErrQueue(func(context.Context, int) error { return nil })
	q.close()

	err := q.Run(t.Context())
	if err == nil {
		t.Fatal("expected error when running a closed queue")
	}
	if err.Error() != "queue is closed" {
		t.Fatalf("unexpected error: %v", err)
	}
}
