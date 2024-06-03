package v1

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/service"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/api/response"
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

// @Summary     Get all expressions
// @Description Получение всех выражений
// @Tags  	    expressions
// @Accept      json
// @Produce     json
// @Success     200 {object} ExpressionsResponse
// @Router       /api/v1/expressions [get]
func (or *OrchestratorResource) GetExpressions(w http.ResponseWriter, r *http.Request) {
	exps := or.expM.GetExpressions()
	const op = "handlers.OrchestratorResource.GetExpression"

	log := or.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	log.Debug("all expressions", slog.Any("expressions", exps))
	//fmt.Println(exps)
	//var respExps
	//respExps := make([]ExpressionData, len(exps))
	respExps := []ExpressionData{}
	//fmt.Println(respExps)

	for _, exp := range exps {
		expD := ExpressionToExpressionData(exp)
		respExps = append(respExps, expD)
	}

	//fmt.Println(respExps)
	log.Debug("all ExpressionToExpressionData", slog.Any("expressions", exps))

	render.JSON(w, r, ExpressionsResponse{Expressions: respExps})

}

// @Summary     Get expression
// @Description Получение выражения по id
// @Tags  	    expressions
// @Accept      json
// @Produce     json
// @Failure     500 {object} response.Response
// @Param       id  path string  true  "get expression by id"
// @Success     200 {object} ExpressionResponse
// @Router       /api/v1/expressions/{id} [get]
func (or *OrchestratorResource) GetExpression(w http.ResponseWriter, r *http.Request) {
	//const op = "handlers.OrchestratorResource.GetExpression"
	//
	//log := or.log.With(
	//	slog.String("op", op),
	//	slog.String("request_id", middleware.GetReqID(r.Context())),
	//)

	id := chi.URLParam(r, "id")

	exp, err := or.expM.GetExpressionById(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, response.Error(err.Error()))
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

// @Summary     Send expression to calculate
// @Description отправить выражения для вычисления
// @Tags  	    expressions
// @Accept      json
// @Produce     json
// @Failure     500 {object} response.Response
// @Param request body CalculateRequest true "request"
// @Success     200 {object} ExpressionData
// @Router       /api/v1/calculate [post]
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
		render.JSON(w, r, response.Error("empty request"))
		return
	}

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		log.Error("invalid request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.ValidationError(validateErr))

		return
	}

	exp := or.expM.AddExpression(req.Expression)

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, ExpressionToExpressionData(exp))
}
