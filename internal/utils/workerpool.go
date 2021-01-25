package utils

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Task is a representation of a task
type Task struct {
	fn   func() error
	errC chan error
}

// WorkerPool a pool of workers able to perform multiple tasks in parallel
type WorkerPool struct {
	workers int
	taskQ   chan Task
	wg      sync.WaitGroup
	sem     *semaphore.Weighted
}

// NewWorkerPool instantiate a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	taskQ := make(chan Task)
	wp := &WorkerPool{
		workers: workers,
		taskQ:   taskQ,
		sem:     semaphore.NewWeighted(int64(workers)),
	}

	wp.wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wp.wg.Done()
			for task := range taskQ {
				task.errC <- task.fn()
				wp.sem.Release(1)
			}
		}()
	}

	return wp
}

// Exec enqueues one task into the worker pool
func (wp *WorkerPool) Exec(f func() error) chan error {
	wp.sem.Acquire(context.Background(), 1)
	t := Task{
		fn:   f,
		errC: make(chan error, 1),
	}
	wp.taskQ <- t
	return t.errC
}

// Close the worker pool
func (wp *WorkerPool) Close() {
	close(wp.taskQ)
	wp.wg.Wait()
}
