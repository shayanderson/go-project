package work

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shayanderson/go-project/v2/internal/test"
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

	test.True(t, th.Allow())
}

func TestThrottler_Allow_ThrottledWithinInterval(t *testing.T) {
	t.Parallel()
	th := NewThrottler(200 * time.Millisecond)

	test.True(t, th.Allow())
	test.False(t, th.Allow())
}

func TestThrottler_Allow_AfterInterval(t *testing.T) {
	t.Parallel()
	th := NewThrottler(5 * time.Millisecond)

	test.True(t, th.Allow())
	time.Sleep(10 * time.Millisecond)
	test.True(t, th.Allow())
}

func TestThrottler_Do(t *testing.T) {
	t.Parallel()
	th := NewThrottler(5 * time.Millisecond)

	var c atomic.Int32

	test.True(t, th.Do(func() {
		c.Add(1)
	}))
	test.Equal(t, 1, c.Load())

	test.False(t, th.Do(func() {
		c.Add(1)
	}))
	test.Equal(t, 1, c.Load())

	time.Sleep(10 * time.Millisecond)

	test.True(t, th.Do(func() {
		c.Add(1)
	}))
	test.Equal(t, 2, c.Load())
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

	test.Equal(t, 1, allowed.Load())
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
