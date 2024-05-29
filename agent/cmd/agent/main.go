package main

import (
	"context"
	"github.com/vingp/DistributedCalculator/agent/config"
	"github.com/vingp/DistributedCalculator/agent/internal/service"
	"github.com/vingp/DistributedCalculator/agent/pkg/logger/handlers/slogpretty"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.Get()

	log := setupLogger(cfg.Env)

	log.Info("app config", slog.Any("config", cfg))

	log.Info(
		"starting agent",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
	log.Debug("debug messages are enabled")

	cal := service.NewCalculator(log)
	ctx, _ := context.WithCancel(context.Background())
	cal.Run(ctx)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
