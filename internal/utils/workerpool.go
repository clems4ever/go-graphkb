package utils

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/semaphore"
)

// ErrTaskAborted error thrown when a taks is aborted
var ErrTaskAborted = errors.New("Aborted")

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
	failed  bool
}

// NewWorkerPool instantiate a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	taskQ := make(chan Task)
	wp := &WorkerPool{
		workers: workers,
		taskQ:   taskQ,
		sem:     semaphore.NewWeighted(int64(workers)),
		failed:  false,
	}

	wp.wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wp.wg.Done()
			for task := range taskQ {
				// If the worker pool has been aborted, we skip all tasks in the queue.
				if wp.failed {
					task.errC <- ErrTaskAborted
					wp.sem.Release(1)
					continue
				}
				err := task.fn()
				wp.sem.Release(1)
				if err != nil {
					wp.failed = true
				}
				task.errC <- err
			}
		}()
	}

	return wp
}

// Exec enqueues one task into the worker pool
func (wp *WorkerPool) Exec(f func() error) chan error {
	errC := make(chan error, 1)
	if wp.failed {
		errC <- ErrTaskAborted
		return errC
	}
	wp.sem.Acquire(context.Background(), 1)
	t := Task{
		fn:   f,
		errC: errC,
	}
	wp.taskQ <- t
	return t.errC
}

// Close the worker pool
func (wp *WorkerPool) Close() {
	close(wp.taskQ)
	wp.wg.Wait()
}
