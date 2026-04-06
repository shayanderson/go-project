package work

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunner(t *testing.T) {
	runner, ctx := NewRunner(t.Context())

	done := make(chan struct{})
	runner.Run(func() error {
		close(done)
		return nil
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not finish")
	}

	select {
	case <-ctx.Done():
		t.Fatal("context canceled before Wait")
	default:
	}

	if err := runner.Wait(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunnerError(t *testing.T) {
	runner, ctx := NewRunner(t.Context())

	errFirst := errors.New("first error")
	errSecond := errors.New("second error")

	runner.Run(func() error {
		return errFirst
	})
	runner.Run(func() error {
		time.Sleep(10 * time.Millisecond)
		return errSecond
	})

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context was not canceled after first error")
	}

	if err := runner.Wait(); !errors.Is(err, errFirst) {
		t.Fatalf("expected %v, got %v", errFirst, err)
	}

	if cause := context.Cause(ctx); !errors.Is(cause, errFirst) {
		t.Fatalf("expected context cause %v, got %v", errFirst, cause)
	}
}

func TestThrottler_Allow_FirstCallAllowed(t *testing.T) {
	t.Parallel()
	th := NewThrottler(100 * time.Millisecond)
	if !th.Allow() {
		t.Fatal("expected first call to Allow() to return true")
	}
}

func TestThrottler_Allow_ThrottledWithinInterval(t *testing.T) {
	t.Parallel()
	th := NewThrottler(200 * time.Millisecond)
	if !th.Allow() {
		t.Fatal("expected first call to Allow() to return true")
	}
	if th.Allow() {
		t.Fatal("expected second call to Allow() within interval to return false")
	}
}

func TestThrottler_Allow_AfterInterval(t *testing.T) {
	t.Parallel()
	th := NewThrottler(5 * time.Millisecond)

	if !th.Allow() {
		t.Fatal("expected first call to Allow() to return true")
	}
	time.Sleep(10 * time.Millisecond)
	if !th.Allow() {
		t.Fatal("expected call to Allow() after interval to return true")
	}
}

func TestThrottler_Do(t *testing.T) {
	t.Parallel()
	th := NewThrottler(5 * time.Millisecond)

	var c atomic.Int32

	if !th.Do(func() {
		c.Add(1)
	}) {
		t.Fatal("expected first call to Do() to return true")
	}
	if c.Load() != 1 {
		t.Fatalf("expected c to be 1, got %d", c.Load())
	}

	if th.Do(func() {
		c.Add(1)
	}) {
		t.Fatal("expected second call to Do() to return false")
	}
	if c.Load() != 1 {
		t.Fatalf("expected c to still be 1, got %d", c.Load())
	}

	time.Sleep(10 * time.Millisecond)

	if !th.Do(func() {
		c.Add(1)
	}) {
		t.Fatal("expected call to Do() after interval to return true")
	}
	if c.Load() != 2 {
		t.Fatalf("expected c to be 2, got %d", c.Load())
	}
}

func TestThrottler_ConcurrentAllow(t *testing.T) {
	t.Parallel()
	th := NewThrottler(200 * time.Millisecond)

	const goroutines = 10

	var wg sync.WaitGroup
	var allowed atomic.Int32

	wg.Add(goroutines)
	for range goroutines {
		wg.Go(func() {
			defer wg.Done()
			if th.Allow() {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()

	if allowed.Load() != 1 {
		t.Fatalf("expected exactly 1 allowed call, got %d", allowed.Load())
	}
}

func TestThrottler_RaceSafety(t *testing.T) {
	t.Parallel()
	th := NewThrottler(10 * time.Millisecond)

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		wg.Go(func() {
			defer wg.Done()
			for range iterations {
				th.Allow()
				time.Sleep(time.Millisecond)
			}
		})
	}

	wg.Wait()
}
