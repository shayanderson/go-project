package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/shayanderson/go-project/app"
	"github.com/shayanderson/go-project/app/config"
)

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

func init() {
	if config.Config.Debug {
		loggerOptions.Level = slog.LevelDebug
	}
	slog.SetDefault(
		slog.New(slog.NewJSONHandler(os.Stdout, loggerOptions)),
	)
}

func main() {
	ctx := context.Background()
	app := app.New()

	if err := app.Run(ctx); err != nil && err != context.Canceled {
		fmt.Printf("app run failed: %v\n", err)
		os.Exit(1)
	}
}
