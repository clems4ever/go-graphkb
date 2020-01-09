package utils

import (
	"time"
)

// RecurrentTask represent a recurrent task
type RecurrentTask struct {
	finishC  chan struct{}
	interval time.Duration
	callback func()

	RunAtStartup bool
}

// NewRecurrentTask create a recurrent task
func NewRecurrentTask(interval time.Duration, callback func()) RecurrentTask {
	return RecurrentTask{
		interval: interval,
		callback: callback,
		finishC:  make(chan struct{}),
	}
}

// Start a recurrent task
func (rt *RecurrentTask) Start() {

	if rt.RunAtStartup {
		rt.callback()
	}

	for {
		select {
		case <-rt.finishC:
			return
		case <-time.After(rt.interval):
			rt.callback()
		}
	}
}

// Stop the recurrent task
func (rt *RecurrentTask) Stop() {
	rt.finishC <- struct{}{}
}
