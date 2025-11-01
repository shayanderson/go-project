package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/shayanderson/go-project/app"
)

// main is the entry point of the application
func main() {
	ctx := context.Background()
	config, err := app.NewConfig()
	if err != nil {
		fatal("failed to create config: %v", err)
	}

	if config.Debug {
		loggerOptions.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, loggerOptions)))

	app, err := app.New(config)
	if err != nil {
		fatal("failed to create app: %v", err)
	}
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		fatal("app run failed: %v", err)
	}
}

// fatal logs a fatal error message and exits the application
func fatal(format string, args ...any) {
	slog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// loggerOptions holds the options for the slog logger
var loggerOptions = &slog.HandlerOptions{
	Level: slog.LevelInfo,
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		// custom time format
		if a.Key == slog.TimeKey {
			t := a.Value.Time()
			a.Value = slog.StringValue(t.Format(time.DateTime))
		}

		return a
	},
}
