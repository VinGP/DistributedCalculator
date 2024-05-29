package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/vingp/DistributedCalculator/orchestrator/pkg/queue"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TaskId string

func NewTaskId(expressionId string, i int) TaskId {
	return TaskId(expressionId + "_" + strconv.Itoa(i) + "_" + uuid.New().String())
}

func (t TaskId) GetExpressionId() string {

	return strings.Split(string(t), "_")[0]
}

func (t TaskId) GetParallelId() int {
	i, _ := strconv.Atoi(strings.Split(string(t), "_")[1])
	return i
}

type TaskManager struct {
	tasks map[TaskId]Task
	queue *queue.ConcurrentQueue[TaskId]
	mutex sync.Mutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[TaskId]Task),
		queue: &queue.ConcurrentQueue[TaskId]{},
		mutex: sync.Mutex{},
	}
}

type Task struct {
	Id            TaskId
	Arg1          string
	Arg2          string
	Operation     string
	OperationTime int
	Completed     bool
	Error         error
	SentAt        time.Time
}

func NewTask(id TaskId, arg1 string, arg2 string, operation string, operationTime int) Task {
	task := Task{
		Id:            id,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     operation,
		OperationTime: operationTime,
	}
	return task
}

func (t *TaskManager) AddTask(task Task) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	id := task.Id
	t.tasks[id] = task
	t.queue.Enqueue(id)
}

func (t *TaskManager) GetTaskExpressionInfo(taskId TaskId) (string, int, error) {
	info := strings.Split(string(taskId), "_")

	expId := info[0]

	i, err := strconv.Atoi(info[1])
	if err != nil {
		return "", 0, err
	}

	return expId, i, nil
}

func (t *TaskManager) DeleteTask(id TaskId) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.tasks, id)
}

//func (t *TaskManager) GetTasks() []Task {
//	return t.tasks
//}

func (t *TaskManager) GetTaskById(id TaskId) (Task, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, ok := t.tasks[id]; !ok {
		return Task{}, errors.New("task not found")
	}

	return t.tasks[id], nil
}

func (t *TaskManager) GetTasks() []Task {
	tasks := make([]Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (t *TaskManager) GetNewTask() (Task, error) {
	var task Task
	var err error
	var notCompletedTasksIds []TaskId

	defer func() {
		for _, tId := range notCompletedTasksIds {
			t.queue.Enqueue(tId)
		}
	}()

	for {
		tId, ok := t.queue.Dequeue()
		if !ok {
			return Task{}, ErrNoTasks
		}

		task, err = t.GetTaskById(tId)

		if err != nil {
			continue
		}

		if task.Completed {
			continue
		}

		if !time.Now().After(task.SentAt.Add(time.Millisecond * time.Duration(task.OperationTime*100))) {
			notCompletedTasksIds = append(notCompletedTasksIds, task.Id)
			continue
		}

		task.SentAt = time.Now()

		t.tasks[tId] = task
		break
	}

	t.queue.Enqueue(task.Id)

	return task, nil
}

func (t *TaskManager) CompleteTask(id TaskId, result float64, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if task, ok := t.tasks[id]; ok {

		task.Completed = true
		task.Error = err

		t.tasks[id] = task
	}
}
