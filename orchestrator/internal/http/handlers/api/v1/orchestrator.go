package v1

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/service"
	resp "github.com/vingp/DistributedCalculator/orchestrator/pkg/api/response"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/logger/sl"
	"io"
	"log/slog"
	"net/http"
)

type OrchestratorResource struct {
	log  *slog.Logger
	expM *service.ExpressionManager
}

func NewOrchestratorResource(log *slog.Logger, expM *service.ExpressionManager) *OrchestratorResource {
	return &OrchestratorResource{log, expM}
}

func (or *OrchestratorResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/expressions", func(r chi.Router) {
		r.Get("/", or.GetExpressions)
		r.Get("/{id}", or.GetExpression)
	})
	r.Post("/calculate", or.PostCalculate)

	return r
}

type ExpressionData struct {
	Id         string `json:"id"`
	Status     string `json:"status"`
	Result     string `json:"result,omitempty"`
	Expression string `json:"expression"`
	Error      string `json:"error,omitempty"`
}

type ExpressionResponse struct {
	Expression ExpressionData `json:"expression"`
}

type ExpressionsResponse struct {
	Expressions []ExpressionData `json:"expressions"`
}

func ExpressionToExpressionData(exp service.Expression) ExpressionData {
	expD := ExpressionData{
		Id:         exp.Id,
		Status:     exp.Status.String(),
		Result:     exp.Result,
		Expression: exp.Expr,
	}
	if exp.Err != nil {
		expD.Error = exp.Err.Error()
	}
	return expD
}

func (or *OrchestratorResource) GetExpressions(w http.ResponseWriter, r *http.Request) {
	exps := or.expM.GetExpressions()
	fmt.Println(exps)
	//var respExps
	//respExps := make([]ExpressionData, len(exps))
	respExps := []ExpressionData{}
	fmt.Println(respExps)

	for _, exp := range exps {
		expD := ExpressionToExpressionData(exp)
		respExps = append(respExps, expD)
	}

	fmt.Println(respExps)

	render.JSON(w, r, ExpressionsResponse{Expressions: respExps})

}

func (or *OrchestratorResource) GetExpression(w http.ResponseWriter, r *http.Request) {
	//const op = "handlers.OrchestratorResource.GetExpression"
	//
	//log := or.log.With(
	//	slog.String("op", op),
	//	slog.String("request_id", middleware.GetReqID(r.Context())),
	//)
	//
	id := chi.URLParam(r, "id")

	exp, err := or.expM.GetExpressionById(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, resp.Error(err.Error()))
		return
	}

	render.JSON(w, r, ExpressionResponse{
		Expression: ExpressionData{
			Id:         exp.Id,
			Status:     exp.Status.String(),
			Result:     exp.Result,
			Expression: exp.Expr,
		},
	},
	)

}

type CalculateRequest struct {
	Expression string `json:"expression" validate:"required"`
}

func (or *OrchestratorResource) PostCalculate(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.OrchestratorResource.PostCalculate"

	log := or.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req CalculateRequest

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error("empty request"))
		return
	}

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		log.Error("invalid request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.ValidationError(validateErr))

		return
	}

	exp := or.expM.AddExpression(req.Expression)

	render.JSON(w, r, ExpressionResponse{Expression: ExpressionToExpressionData(exp)})
}
