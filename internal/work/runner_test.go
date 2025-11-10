package work

import (
	"context"
	"errors"
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
