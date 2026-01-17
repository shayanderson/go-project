package work

import (
	"context"
	"sync"
	"time"
)

// Runner is a task Runner
type Runner struct {
	cancel  func(error)
	err     error
	errOnce sync.Once
	wg      sync.WaitGroup
}

// NewRunner creates a new Runner
func NewRunner(ctx context.Context) (*Runner, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Runner{cancel: cancel}, ctx
}

// Run runs a function and handles errors
// sets the first error to the app error
func (g *Runner) Run(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := fn(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
		}
	}()
}

// Wait blocks until all app goroutines are done
// returns the first error if exists
func (g *Runner) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// Throttler limits how often an action may run by allowing
// at most one execution per interval
type Throttler struct {
	interval time.Duration
	lastRun  time.Time
	mu       sync.Mutex
}

// NewThrottler creates a new Throttler with the specified interval
func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{interval: interval}
}

// Allow reports whether an action may run now
// returns true if the caller is considered to have "used" the interval
func (t *Throttler) Allow() bool {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.lastRun.IsZero() && now.Sub(t.lastRun) < t.interval {
		return false
	}
	t.lastRun = now
	return true
}

// Do executes the provided function if allowed by the throttler
// returns true if the function was executed
func (t *Throttler) Do(fn func()) bool {
	if t.Allow() {
		fn()
		return true
	}
	return false
}
