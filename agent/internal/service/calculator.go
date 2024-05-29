package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/vingp/DistributedCalculator/agent/config"
	"github.com/vingp/DistributedCalculator/agent/pkg/logger/sl"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Calculator struct {
	log *slog.Logger
}

func NewCalculator(log *slog.Logger) *Calculator {
	return &Calculator{log: log}
}

type TaskData struct {
	Id            string `json:"id,required"`
	Arg1          string `json:"arg1,required"`
	Arg2          string `json:"arg2,required"`
	Operation     string `json:"operation,required"`
	OperationTime int    `json:"operation_time,required"`
}

type OrchestratorTaskDataResp struct {
	Task TaskData `json:"task,required"`
}
type TaskResult struct {
	Id     string  `json:"id,required"`
	Result float64 `json:"result,required"`
	Error  string  `json:"error,omitempty"`
}

func (c *Calculator) GetTask() (TaskData, error) {
	resp, err := http.Get(config.Get().OrchestratorURL + "/internal/task")
	if err != nil {
		return TaskData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var td OrchestratorTaskDataResp

		err := render.DecodeJSON(resp.Body, &td)
		fmt.Println("td", td)
		if errors.Is(err, io.EOF) {
			c.log.Error("request body is empty")

			return TaskData{}, err
		}

		if err != nil {
			c.log.Error("GetTask err", sl.Err(err))
			return TaskData{}, err
		}
		//
		//data, _ := io.ReadAll(resp.Body)
		return td.Task, nil
	} else if resp.StatusCode == 404 {
		return TaskData{}, errors.New("no task")
	} else {
		return TaskData{}, errors.New("server error")
	}

}

func (c *Calculator) SentDownTask(tr TaskResult) {
	data, err := json.Marshal(tr)
	if err != nil {
		c.log.Info("Error sent task res", sl.Err(err))
		return
	}
	//
	//io.Writer()
	_, err = http.Post(config.Get().OrchestratorURL+"/internal/task", "application/json", bytes.NewReader(data))
	if err != nil {
		c.log.Error("err SentDownTask", sl.Err(err))
	}
	//fmt.Println("Sent Task", tr)
}

func (c *Calculator) Run(ctx context.Context) {
	in := make(chan TaskData)
	out := make(chan TaskResult)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(in)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(500 * time.Millisecond)
				newTask, err := c.GetTask()
				if err != nil {
					continue
				}
				in <- newTask
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()

		for {
			select {
			case ans := <-out:
				fmt.Println(ans)
				c.SentDownTask(ans)
			case <-ctx.Done():
				return
			}
		}

	}()
	//ctx, cancel := context.WithCancel(context.Background())
	go Test(config.Get().ComputingPower, ctx, in, out)
	wg.Wait()
}

func Test(count int, ctx context.Context, in chan TaskData, out chan TaskResult) {
	wg := sync.WaitGroup{} // для ожидания завершения

	for i := 0; i < count; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case task := <-in:
					fmt.Println("New task", task)
					out <- calculate(task)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	wg.Wait()
}

func calculate(td TaskData) TaskResult {
	fmt.Println(td)
	res := TaskResult{
		Id:     td.Id,
		Result: 0,
		Error:  "",
	}
	op1, err := strconv.ParseFloat(td.Arg1, 64)

	if err != nil {
		res.Error = "invalid number"
		return res
	}
	op2, err := strconv.ParseFloat(td.Arg2, 64)
	if err != nil {
		res.Error = "invalid number"
		return res
	}

	switch td.Operation {
	case "+":
		res.Result = op1 + op2
	case "-":
		res.Result = op1 - op2
	case "*":
		res.Result = op1 * op2
	case "/":
		if op2 == 0 {
			res.Error = "division by zero"
			return res
		}
		res.Result = op1 / op2
	case "^":
		res.Result = math.Pow(op1, op2)

	default:
		res.Error = "invalid operation"
	}
	time.Sleep(time.Duration(td.OperationTime) * time.Millisecond)
	return res
}
