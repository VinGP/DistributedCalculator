package queue

import "sync"

type ConcurrentQueue[T any] struct {
	queue []T // здесь хранить элементы очереди
	mutex sync.Mutex
}

func (c *ConcurrentQueue[T]) Enqueue(element T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.queue = append(c.queue, element)
}

func (c *ConcurrentQueue[T]) Dequeue() (T, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.queue) == 0 {
		var res T
		return res, false
	}
	res := c.queue[0]
	c.queue = c.queue[1:]
	return res, true
}
