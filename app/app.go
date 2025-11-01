package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/shayanderson/go-project/app/handler"
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

	srv := newServer(ServerOptions{
		Addr:              a.config.HTTPServerAddr,
		ReadHeaderTimeout: a.config.HTTPServerReadHeaderTimeout,
		ReadTimeout:       a.config.HTTPServerReadTimeout,
		WriteTimeout:      a.config.HTTPServerWriteTimeout,
	})

	// example middleware: logging
	srv.Use(func(next Handler) Handler {
		return Handler(func(w http.ResponseWriter, r *http.Request) error {
			slog.Info("[http] handling request", "method", r.Method, "url", r.URL.String())
			return next(w, r)
		})
	})

	rootHandler := handler.NewRoot()
	srv.Handle(httpRootPattern, rootHandler.Index, func(next Handler) Handler {
		return Handler(func(w http.ResponseWriter, r *http.Request) error {
			slog.Info("[http] root handler middleware")
			return next(w, r)
		})
	})
	srv.Handle("/", rootHandler.NotFound) // catch-all for not found

	runner.run(func() error {
		if err := srv.Start(); err != nil {
			return fmt.Errorf("http server start failed: %w", err)
		}
		return nil
	})

	runner.run(func() error {
		<-ctx.Done()
		if err := srv.Stop(); err != nil {
			return fmt.Errorf("http server stop failed: %w", err)
		}
		return nil
	})

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
