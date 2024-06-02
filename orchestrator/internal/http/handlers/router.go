package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	docs "github.com/vingp/DistributedCalculator/orchestrator/docs"
	apiV1 "github.com/vingp/DistributedCalculator/orchestrator/internal/http/handlers/api/v1"
	apiInternal "github.com/vingp/DistributedCalculator/orchestrator/internal/http/handlers/internal_api"
	mwLogger "github.com/vingp/DistributedCalculator/orchestrator/internal/http/middleware/logger"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/service"
	"net/http"

	"log/slog"
)

// @title           Distributed calculator API
// @version         1.0
// @description     This is a distributed arithmetic expression calculator api

func NewRouter(r chi.Router, log *slog.Logger, tM *service.TaskManager, expM *service.ExpressionManager) {
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

	//r.Mount("/swagger", httpSwagger.WrapHandler)
	//r.Mount("/swagger/*", httpSwagger.WrapHandler(swaggerfiles.Handler))
	//r.Get("/swagger/*", httpSwagger.Handler(
	//	docs.SwaggerInfo.Host = context.Request.Host
	//	httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
	//))
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		baseURL := r.Host
		docs.SwaggerInfo.Host = baseURL
		httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
		).ServeHTTP(w, r)
	})

	//tM := service.NewTaskManager()
	//expM := service.NewExpressionManager(tM)
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
}
