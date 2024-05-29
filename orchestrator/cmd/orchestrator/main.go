package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vingp/DistributedCalculator/orchestrator/config"
	apiV1 "github.com/vingp/DistributedCalculator/orchestrator/internal/http/handlers/api/v1"
	apiInternal "github.com/vingp/DistributedCalculator/orchestrator/internal/http/handlers/internal_api"
	mwLogger "github.com/vingp/DistributedCalculator/orchestrator/internal/http/middleware/logger"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/service"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/logger/handlers/slogpretty"
	"log/slog"
	"net/http"
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
		"starting orchestrator",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)
	log.Debug("debug messages are enabled")

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	tM := service.NewTaskManager()
	expM := service.NewExpressionManager(tM)
	//r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("welcome"))
	//})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Mount("/", apiV1.NewOrchestratorResource(log, expM).Routes())
		})
	})

	r.Route("/internal", func(r chi.Router) {
		r.Mount("/", apiInternal.NewTaskResource(log, tM, expM).Routes())
	})
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Info("server started", slog.String("addr", addr))
	err := http.ListenAndServe(addr, r)

	if err != nil {
		log.Error("failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
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
