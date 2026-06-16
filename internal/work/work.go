package work

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Accumulator accumulates values and invokes a callback when either:
//   - the total reaches or exceeds max
//   - delay elapses and total > 0
//
// after either condition, the total is reset to zero
type Accumulator interface {
	// Add adds n to the accumulator
	Add(n int)
	// Close stops the accumulator and releases resources
	Close()
}

// accumulator implements Accumulator
type accumulator struct {
	closed bool
	delay  time.Duration
	fn     func(total int)
	max    int
	mu     sync.Mutex
	timer  *time.Timer
	total  int
}

// NewAccumulator creates a new Accumulator
func NewAccumulator(
	ctx context.Context,
	delay time.Duration,
	max int,
	fn func(total int),
) (Accumulator, error) {
	if delay <= 0 {
		return nil, errors.New("delay must be greater than zero")
	}
	if max < 1 {
		return nil, errors.New("max must be greater than zero")
	}
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}

	a := &accumulator{
		delay: delay,
		fn:    fn,
		max:   max,
		timer: time.NewTimer(delay),
	}

	go a.run(ctx)

	return a, nil
}

// Add adds n to the accumulator
func (a *accumulator) Add(n int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}

	a.total += n

	if a.total >= a.max {
		a.flushLocked()
		return
	}

	// restart timer
	if !a.timer.Stop() {
		select {
		case <-a.timer.C:
		default:
		}
	}
	a.timer.Reset(a.delay)
}

// Close flushes any remaining accumulated value and stops the accumulator
// Close is safe to call multiple times
func (a *accumulator) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}
	a.closed = true

	// stop timer and drain channel if needed
	if !a.timer.Stop() {
		select {
		case <-a.timer.C:
		default:
		}
	}

	if a.total > 0 {
		total := a.total
		a.total = 0
		a.fn(total)
	}
}

// flushLocked invokes the callback and resets the accumulator
// caller must hold a.mu
func (a *accumulator) flushLocked() {
	if a.total == 0 {
		return
	}

	total := a.total
	a.total = 0

	a.fn(total)
}

// run handles delayed flushes and context cancellation
func (a *accumulator) run(ctx context.Context) {
	defer a.timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-a.timer.C:
			a.mu.Lock()

			if a.total > 0 {
				a.flushLocked()
			} else {
				a.timer.Reset(a.delay)
			}

			a.mu.Unlock()
		}
	}
}

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
