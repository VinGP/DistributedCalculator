package orchestrator

import (
	"github.com/go-chi/chi/v5"
	"github.com/vingp/DistributedCalculator/orchestrator/config"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/http/handlers"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/service"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/storage/sqlite"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/httpserver"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/logger"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/logger/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) {
	log := logger.NewLogger(cfg.Env)

	log.Info("app config", slog.Any("config", cfg))

	log.Info(
		"starting orchestrator",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("app - Run - sqlite.New: %w", sl.Err(err))
	}
	defer func(storage *sqlite.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error("app - Run - storage.Close: %w", sl.Err(err))
		}
	}(storage)

	tM := service.NewTaskManager()
	expM := service.NewExpressionManager(tM)

	r := chi.NewRouter()

	handlers.NewRouter(r, log, tM, expM)

	httpServer := httpserver.New(r, httpserver.Port(cfg.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: ", slog.Any("signal", s.String()))
	case err := <-httpServer.Notify():
		log.Error("app - Run - httpServer.Notify: %w", sl.Err(err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		log.Error("app - Run - httpServer.Shutdown: %w", sl.Err(err))
	}

}
