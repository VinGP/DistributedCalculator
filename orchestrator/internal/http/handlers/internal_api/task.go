package internal_api

import (
	"errors"
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

type TaskResource struct {
	log *slog.Logger
	tm  *service.TaskManager
	em  *service.ExpressionManager
}

func NewTaskResource(log *slog.Logger, tm *service.TaskManager, em *service.ExpressionManager) *TaskResource {
	return &TaskResource{log, tm, em}
}

func (tr *TaskResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/task", func(r chi.Router) {
		r.Get("/", tr.GetTask)
		r.Post("/", tr.PostTask)
	})

	return r
}

type TaskData struct {
	Id            string `json:"id"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Operation     string `json:"operation"`
	OperationTime int    `json:"operation_time"`
}

type TaskResponse struct {
	Task TaskData `json:"task"`
}

// @Summary     Get task
// @Description получить задачу на выполнение
// @Tags  	    tasks
// @Accept      json
// @Produce     json
// @Failure     500 {object} response.Response
// @Success     200 {object} TaskResponse
// @Router       /internal/task [get]
func (tr *TaskResource) GetTask(w http.ResponseWriter, r *http.Request) {
	task, err := tr.tm.GetNewTask()

	if errors.Is(err, service.ErrNoTasks) {
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, resp.Error(err.Error()))

		return
	}

	taskData := TaskData{
		Id:            string(task.Id),
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: task.OperationTime,
	}
	render.JSON(w, r, TaskResponse{taskData})
}

type TaskDoneRequest struct {
	Id     string  `json:"id,required"`
	Result float64 `json:"result,required"`
	Error  string  `json:"error,omitempty"`
}

// @Summary     Post task
// @Description отправить результат выполнения задачи
// @Tags  	    tasks
// @Accept      json
// @Produce     json
// @Failure     500 {object} response.Response
// @Param request body TaskDoneRequest true "request"
// @Success     200 {object} TaskResponse
// @Router       /internal/task [post]
func (tr *TaskResource) PostTask(w http.ResponseWriter, r *http.Request) {
	const op = "internal.http.internal_api.TaskResource.PostTask"

	log := tr.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req TaskDoneRequest

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, resp.Error("empty request"))
		return
	}

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, resp.Error("failed to decode request"))
		return
	}

	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		log.Error("invalid request", sl.Err(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, resp.ValidationError(validateErr))

		return
	}

	if req.Error != "" {
		err = errors.New(req.Error)
	}
	err = tr.em.UpdateExpressionTasks(service.TaskResult{
		Id:     service.TaskId(req.Id),
		Result: req.Result,
		Err:    err,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(err.Error()))
		return
	}

	render.JSON(w, r, make(map[string]string))
}
