package work

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewAccumulator_InvalidArgs(t *testing.T) {
	ctx := t.Context()

	_, err := NewAccumulator(ctx, 0, 1, func(int) {})
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = NewAccumulator(ctx, time.Second, -1, func(int) {})
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = NewAccumulator(ctx, time.Second, 1, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = NewAccumulator(ctx, time.Second, 0, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAccumulator_Max(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var total int

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		10,
		func(n int) {
			total = n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(3)
	a.Add(4)

	if total != 0 {
		t.Fatalf("got %d, want 0", total)
	}

	a.Add(3)

	if total != 10 {
		t.Fatalf("got %d, want 10", total)
	}
}

func TestAccumulator_Delay(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ch := make(chan int, 1)

	a, err := NewAccumulator(
		ctx,
		50*time.Millisecond,
		100,
		func(n int) {
			ch <- n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(3)
	a.Add(4)

	select {
	case n := <-ch:
		if n != 7 {
			t.Fatalf("got %d, want 7", n)
		}

	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback")
	}
}

func TestAccumulator_Close(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var total int

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		100,
		func(n int) {
			total = n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(2)
	a.Add(5)

	a.Close()

	if total != 7 {
		t.Fatalf("got %d, want 7", total)
	}
}

func TestAccumulator_CloseEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	calls := 0

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		100,
		func(int) {
			calls++
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Close()

	if calls != 0 {
		t.Fatalf("got %d calls, want 0", calls)
	}
}

func TestAccumulator_CloseTwice(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	calls := 0

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		100,
		func(int) {
			calls++
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(5)

	a.Close()
	a.Close()

	if calls != 1 {
		t.Fatalf("got %d calls, want 1", calls)
	}
}

func TestAccumulator_AddAfterClose(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	calls := 0

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		10,
		func(n int) {
			calls++
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(5)
	a.Close()

	if calls != 1 {
		t.Fatalf("got %d calls, want 1", calls)
	}

	a.Add(5)

	if calls != 1 {
		t.Fatalf("got %d calls, want 1", calls)
	}
}

func TestAccumulator_FlushLockedZero(t *testing.T) {
	a := &accumulator{
		fn: func(int) {
			t.Fatal("should not be called")
		},
	}

	a.mu.Lock()
	a.flushLocked()
	a.mu.Unlock()
}

func TestAccumulator_TimerResets(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ch := make(chan int, 1)

	a, err := NewAccumulator(
		ctx,
		50*time.Millisecond,
		100,
		func(n int) {
			ch <- n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(3)

	// wait for half the delay, then add again
	// this should restart the timer
	time.Sleep(25 * time.Millisecond)

	a.Add(4)

	// ensure the callback does NOT happen at the original 50ms mark
	select {
	case n := <-ch:
		t.Fatalf("callback fired too early with %d", n)

	case <-time.After(35 * time.Millisecond):
		// good: callback has not fired yet
	}

	// it should fire after the delay from the second Add()
	select {
	case n := <-ch:
		if n != 7 {
			t.Fatalf("got %d, want 7", n)
		}

	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback")
	}
}

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

func TestAccumulator_NoMax(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ch := make(chan int, 1)

	a, err := NewAccumulator(
		ctx,
		50*time.Millisecond,
		0, // no threshold flushing
		func(n int) {
			ch <- n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(3)
	a.Add(4)
	a.Add(5)

	// should not flush immediately
	select {
	case n := <-ch:
		t.Fatalf("callback fired too early with %d", n)

	default:
	}

	// should flush after delay
	select {
	case n := <-ch:
		if n != 12 {
			t.Fatalf("got %d, want 12", n)
		}

	case <-time.After(time.Second):
		t.Fatal("timeout waiting for callback")
	}
}

func TestAccumulator_NoMaxClose(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var total int

	a, err := NewAccumulator(
		ctx,
		time.Hour,
		0,
		func(n int) {
			total = n
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(2)
	a.Add(3)

	a.Close()

	if total != 5 {
		t.Fatalf("got %d, want 5", total)
	}
}

func TestAccumulator_EmptyTimerReset(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	calls := 0

	_, err := NewAccumulator(
		ctx,
		10*time.Millisecond,
		0,
		func(int) {
			calls++
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// let several timer expirations happen with total == 0
	time.Sleep(50 * time.Millisecond)

	if calls != 0 {
		t.Fatalf("got %d calls, want 0", calls)
	}
}

func TestAccumulator_CloseAfterTimerFired(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan struct{}, 1)

	a, err := NewAccumulator(
		ctx,
		10*time.Millisecond,
		0,
		func(int) {
			done <- struct{}{}
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(1)

	<-done

	// timer has already fired and been drained by run()
	a.Close()
}
