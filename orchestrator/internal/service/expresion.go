package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vingp/DistributedCalculator/orchestrator/config"
	"regexp"
	"sort"
	"sync"
	"time"
)

type ExpressionStatus int

func (es ExpressionStatus) String() string {
	switch es {
	case Created:
		return "Created"
	case Started:
		return "Started"
	case Finished:
		return "Finished"
	default:
		return "Unknown"
	}
}

const (
	Created ExpressionStatus = iota
	Started
	Finished
)

type TaskResult struct {
	Id     TaskId  `json:"id,omitempty"`
	Result float64 `json:"result,omitempty"`
	Err    error   `json:"err,omitempty"`
}

type Expression struct {
	Id          string
	Expr        string
	Result      string
	Tokens      []Token
	Err         error
	Status      ExpressionStatus
	RPN         []Token
	TaskIds     []TaskId
	TasksResult []TaskResult
	CreatAt     time.Time

	prevRPNLen  int
	taskManager *TaskManager
	tokenizer   *Tokenizer
}

func checkBrackets(tokens []Token) error {
	var stack []Token
	for _, token := range tokens {
		if token.Type == LeftBracket {
			stack = append(stack, token)
		} else if token.Type == RightBracket {
			if len(stack) == 0 {

				return errors.New("invalid brackets in expression")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		return errors.New("invalid brackets in expression")
	}

	for i, token := range tokens {
		if token.Type == RightBracket && tokens[i-1].Type == LeftBracket {
			return errors.New("invalid brackets in expression")
		}
	}

	return nil
}

func checkOperands(expression string) bool {
	re := regexp.MustCompile(`^[\d*+/\-()^.\s]+$`)
	return re.MatchString(expression)
}

func checkExpression(exp string, tokens []Token) error {
	if err := checkBrackets(tokens); err != nil {
		return err
	}

	if !checkOperands(exp) {
		return errors.New("invalid expression")
	}
	return nil

}

func opPriority(op TokenType) int {
	switch op {
	case Plus, Minus:
		return 1
	case Multiply, Divide:
		return 2
	case Power:
		return 3
	default:
		return 0
	}
}
func NewExpression(expr string, taskManager *TaskManager) *Expression {
	id := uuid.New().String()

	exp := &Expression{
		Id:      id,
		Expr:    expr,
		Tokens:  []Token{},
		Err:     nil,
		Status:  Created,
		CreatAt: time.Now(),

		taskManager: taskManager,
		tokenizer:   NewTokenizer(),
	}

	tokens, err := exp.tokenizer.Tokenize(expr)

	if err != nil {
		exp.Finish(err)

		return exp
	}

	exp.Tokens = tokens

	if err := checkExpression(expr, tokens); err != nil {
		exp.Finish(err)
		return exp
	}

	exp.RPN = exp.ToRPN(tokens)
	return exp
}

func (e *Expression) ToRPN(tokens []Token) []Token {
	var result []Token
	var stack []Token

	for _, token := range tokens {
		switch token.Type {
		case LeftBracket:
			stack = append(stack, token)
		case RightBracket:
			for len(stack) > 0 && stack[len(stack)-1].Type != LeftBracket {
				result = append(result, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		case Plus, Minus, Multiply, Divide, Power:
			for len(stack) > 0 && opPriority(token.Type) <= opPriority(stack[len(stack)-1].Type) {
				result = append(result, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		default:
			result = append(result, token)
		}
	}

	for len(stack) > 0 {
		result = append(result, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return result
}

func (e *Expression) GetParallelOperations(tokens []Token) map[int][]Token {
	var res map[int][]Token

	res = make(map[int][]Token)

	for i := 0; i < len(tokens)-2; i++ {
		n1, n2, op := tokens[i], tokens[i+1], tokens[i+2]
		if n1.Type == Number && n2.Type == Number && op.Type != Number {
			res[i] = []Token{n1, n2, op}
		}
	}

	return res
}

func (e *Expression) AllTasksFinished() bool {
	if len(e.TaskIds) == len(e.TasksResult) {
		return false
	}
	return true
}

func (e *Expression) Finish(err error) {

	e.Status = Finished
	e.Err = err
	e.TaskIds = []TaskId{}
	e.TasksResult = []TaskResult{}

}

func (e *Expression) SimplifyRPN() {

	if len(e.TasksResult) != len(e.TaskIds) {
		return
	}

	var offset int
	fmt.Println("TasksResult1", e.TasksResult)

	sort.Slice(e.TasksResult, func(i, j int) bool {
		n := e.TasksResult[i].Id.GetParallelId()
		m := e.TasksResult[j].Id.GetParallelId()
		return n < m
	})

	fmt.Println("TasksResult", e.TasksResult)

	for _, op := range e.TasksResult {
		i := op.Id.GetParallelId()
		//fmt.Println(i, offset)
		//fmt.Println(op)
		e.RPN[i-offset].Value = fmt.Sprintf("%v", op.Result)
		e.RPN = append(e.RPN[:i-offset+1], e.RPN[i-offset+3:]...)
		offset += 2
	}

	if e.prevRPNLen == len(e.RPN) {
		e.Finish(errors.New("failed calculation"))
		return
	}

	e.TasksResult = []TaskResult{}
	e.TaskIds = []TaskId{}

	e.prevRPNLen = len(e.RPN)
	if len(e.RPN) == 1 {
		e.Finish(nil)
		e.Result = e.RPN[0].Value
		return
	}

	e.UpdateTasks()

	//if
}

func GetOperationTime(op string) int {
	cfg := config.Get()
	switch op {
	case "+":
		return cfg.TimeAdditionMs
	case "-":
		return cfg.TimeSubtractionMs
	case "/":
		return cfg.TimeDivisionsMs
	case "*":
		return cfg.TimeMultiplicationsMs
	case "^":
		return cfg.TimeSubtractionMs
	default:
		return 1000
	}
}

func (e *Expression) UpdateTasks() {
	if e.Status == Finished {
		return
	}
	if len(e.RPN) == 1 {
		e.Finish(nil)
		e.Result = e.RPN[0].Value
		return
	}

	op := e.GetParallelOperations(e.RPN)
	//fmt.Println("op", op)
	if len(op) == 0 {
		e.Finish(errors.New("invalid expression"))
	}

	var tasks []Task
	for i, v := range op {
		taskId := NewTaskId(e.Id, i)

		task := NewTask(taskId, v[0].Value, v[1].Value, v[2].Value, GetOperationTime(v[2].Value))
		//fmt.Println("task", task)
		e.taskManager.AddTask(task)
		tasks = append(tasks, task)

		e.TaskIds = append(e.TaskIds, taskId)
	}

}

func (e *Expression) Start() {
	e.Status = Started
	e.UpdateTasks()
}

func (e *Expression) AddTaskResult(tr TaskResult) {
	fmt.Println("AddTaskResult1")
	if tr.Err != nil {
		e.Finish(tr.Err)
		return
	}
	fmt.Println("AddTaskResult2")

	e.TasksResult = append(e.TasksResult, tr)
	e.taskManager.CompleteTask(tr.Id, tr.Result, tr.Err)
	fmt.Println("SimplifyRPN")
	e.SimplifyRPN()
}

func (e *Expression) String() string {
	return "Expression: " + e.Expr + " Status: " + ExpressionStatusToString(e.Status) + " Result: " + fmt.Sprint(e.Result) + " Err: " + fmt.Sprint(e.Err) + " RPN: " + fmt.Sprint(e.RPN) + " TaskIds: " + fmt.Sprint(e.TaskIds) + " TasksResult: " + fmt.Sprint(e.TasksResult) + "\n"
}

func ExpressionStatusToString(status ExpressionStatus) string {
	switch status {
	case Created:
		return "Created"
	case Started:
		return "Started"
	case Finished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type ExpressionManager struct {
	expressions map[string]*Expression
	mutex       sync.Mutex
	taskManager *TaskManager
}

func NewExpressionManager(taskManager *TaskManager) *ExpressionManager {
	return &ExpressionManager{
		expressions: make(map[string]*Expression),
		taskManager: taskManager,

		mutex: sync.Mutex{},
	}
}

func (em *ExpressionManager) AddExpression(expr string) Expression {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	exp := NewExpression(expr, em.taskManager)
	em.expressions[exp.Id] = exp
	exp.Start()
	return *exp
}

func (em *ExpressionManager) GetExpressionById(expId string) (*Expression, error) {
	//em.mutex.Lock()
	//defer em.mutex.Unlock()
	exp, ok := em.expressions[expId]
	if !ok {
		return &Expression{}, ErrExpressionNotFound
	}
	return exp, nil
}

func (em *ExpressionManager) UpdateExpressionTasks(tr TaskResult) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	_, err := em.taskManager.GetTaskById(tr.Id)
	if err != nil {
		return err
	}

	expId := tr.Id.GetExpressionId()
	exp, err := em.GetExpressionById(expId)
	if err != nil {
		return err
	}

	exp.AddTaskResult(tr)
	em.expressions[expId] = exp
	return nil
}

func (em *ExpressionManager) GetExpressions() []Expression {
	fmt.Println(em.expressions)
	exps := []Expression{}
	for _, v := range em.expressions {
		exps = append(exps, *v)
	}

	sort.Slice(exps, func(i, j int) bool {
		n := exps[i].CreatAt
		m := exps[j].CreatAt
		return n.After(m)
	})

	return exps
}
