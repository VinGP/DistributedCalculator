package entities

import "time"

type Token struct {
	Type  string
	Value string
}

type ExpressionToken struct {
	Token
	Id           int
	ExpressionId string
	Pos          int
}

type RPNToken struct {
	Type         string
	Value        string
	Id           string
	ExpressionId int
	Pos          int
}

type TaskStatus int

const (
	TaskCreated TaskStatus = iota
	TaskStarted
	TaskFinished
)

func GetTaskStatus(status string) TaskStatus {
	switch status {
	case "Created":
		return TaskCreated
	case "Started":
		return TaskStarted
	case "Finished":
		return TaskFinished
	default:
		return TaskCreated
	}
}

func (es TaskStatus) String() string {
	switch es {
	case TaskCreated:
		return "Created"
	case TaskStarted:
		return "Started"
	case TaskFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type Task struct {
	Id           string
	ExpressionId string
	Status       TaskStatus

	Arg1          string
	Arg2          string
	Operation     string
	OperationTime int

	TokensPos []int
	StartedAt time.Time
}

type TaskData struct {
	Id            string `json:"id"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Operation     string `json:"operation"`
	OperationTime int    `json:"operation_time"`
}

type ExpressionStatus int

const (
	Created ExpressionStatus = iota
	Started
	Finished
)

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

type Expression struct {
	Id         int
	UserId     int
	Expr       string
	Result     string
	Err        string
	Status     ExpressionStatus
	RPN        []*RPNToken
	CreatAt    time.Time
	prevRPNLen int
}
