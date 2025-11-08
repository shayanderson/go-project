package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/shayanderson/go-project/v2/entity"
	"github.com/shayanderson/go-project/v2/infra/cache"
	"github.com/shayanderson/go-project/v2/internal/server"
	"github.com/shayanderson/go-project/v2/service"
)

// App is the main application
type App struct {
	config Config
}

// New creates a new App instance
func New(config Config) (*App, error) {
	return &App{config: config}, nil
}

// Run runs the application
func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	runner, ctx := newRunner(ctx)

	// set global http bind limit
	server.LimitReadSize = a.config.HTTPBindLimitReadSize

	// create http server
	srv := server.New(server.Options{
		Addr:              a.config.HTTPServerAddr,
		ReadHeaderTimeout: a.config.HTTPServerReadHeaderTimeout,
		ReadTimeout:       a.config.HTTPServerReadTimeout,
		WriteTimeout:      a.config.HTTPServerWriteTimeout,
	})

	// create api service
	api := service.NewAPI(srv, service.Infra{
		// inject item store
		// normally this would be a database or persistent store
		// like: `infra/db/item.go` and use `db.NewItem(...)`
		ItemStore: cache.New[entity.Item, int](),
	})

	// start api server
	runner.run(func() error {
		if err := api.Start(); err != nil {
			return fmt.Errorf("http server start failed: %w", err)
		}
		return nil
	})

	// handle shutdown
	runner.run(func() error {
		<-ctx.Done()
		if err := api.Stop(); err != nil {
			return fmt.Errorf("http server stop failed: %w", err)
		}
		return nil
	})

	// wait for all tasks to complete
	// in this case, wait for api/http server to stop
	return runner.wait()
}

// runner is a task runner
type runner struct {
	cancel  func(error)
	err     error
	errOnce sync.Once
	wg      sync.WaitGroup
}

// newRunner creates a new runner
func newRunner(ctx context.Context) (*runner, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &runner{cancel: cancel}, ctx
}

// run runs a function and handles errors
// sets the first error to the app error
func (g *runner) run(fn func() error) {
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

// wait blocks until all app goroutines are done
// returns the first error if exists
func (g *runner) wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}
