package main

import (
	"github.com/vingp/DistributedCalculator/orchestrator/config"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/app/orchestrator"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.Get()

	orchestrator.Run(cfg)

	//log := setupLogger(cfg.Env)
	//
	//
	//
	//r := chi.NewRouter()
	//
	//addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	//log.Info("server started", slog.String("addr", addr))
	//err := http.ListenAndServe(addr, r)
	//
	//if err != nil {
	//	log.Error("failed to start server", slog.String("error", err.Error()))
	//	os.Exit(1)
	//}
}
