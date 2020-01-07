package utils

import (
	"time"
)

// RecurrentTask represent a recurrent task
type RecurrentTask struct {
	cancelChannel chan bool
	interval      time.Duration
	callback      func()

	RunAtStartup bool
}

// NewRecurrentTask create a recurrent task
func NewRecurrentTask(interval time.Duration, callback func()) RecurrentTask {
	return RecurrentTask{
		cancelChannel: make(chan bool),
		interval:      interval,
		callback:      callback,
	}
}

// Start a recurrent task
func (rt *RecurrentTask) Start() {
	go func() {
		if rt.RunAtStartup {
			rt.callback()
		}

		for {
			select {
			case <-rt.cancelChannel:
				break
			case <-time.After(rt.interval):
				rt.callback()
			}
		}
	}()
}

// Stop the recurrent task
func (rt *RecurrentTask) Stop() {
	rt.cancelChannel <- true
}
