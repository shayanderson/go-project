package work

import (
	"context"
	"sync"
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
