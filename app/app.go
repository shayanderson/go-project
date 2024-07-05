package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shayanderson/go-project/app/config"
	"github.com/shayanderson/go-project/app/server"
	"github.com/shayanderson/go-project/app/server/handler"
	"github.com/shayanderson/go-project/app/server/middleware"
)

// App is the main application
type App struct {
	cancel  func(error)
	err     error
	errOnce sync.Once
	wg      sync.WaitGroup
}

// New creates a new App
func New() *App {
	return &App{}
}

// init initializes the app
func (a *App) init(ctx context.Context) error {
	return nil
}

// run runs a function and handles errors
// sets the first error to the app error
func (a *App) run(fn func() error) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		if err := fn(); err != nil {
			a.errOnce.Do(func() {
				a.err = err
				if a.cancel != nil {
					a.cancel(a.err)
				}
			})
		}
	}()
}

// Run runs the app
func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.init(ctx); err != nil {
		return fmt.Errorf("app init failed: %w", err)
	}

	ctx, a.cancel = context.WithCancelCause(ctx)

	// http server
	srv := server.New(config.Config.ServerPort)

	// http middleware
	srv.Router.Use(server.LoggerMiddleware)
	srv.Router.Use(server.RecoverMiddleware)
	srv.Router.Use(middleware.TestMiddleware)

	// http handlers
	testHandler := handler.NewTestHandler()

	// http routes
	srv.Router.Get("/test", testHandler.Get, middleware.TestHandlerMiddleware)
	srv.Router.Get("/test/{name}", testHandler.GetEchoName)

	a.run(srv.Start)
	a.run(func() error {
		<-ctx.Done()
		return srv.Stop(ctx)
	})

	return a.wait()
}

// wait blocks until all app goroutines are done
// returns the first error if exists
func (a *App) wait() error {
	a.wg.Wait()
	if a.cancel != nil {
		a.cancel(a.err)
	}
	return a.err
}
