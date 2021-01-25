package utils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPoolNotBlocking(t *testing.T) {
	p := NewWorkerPool(5)
	defer p.Close()

	f1 := p.Exec(func() error {
		return nil
	})
	f2 := p.Exec(func() error {
		return nil
	})

	<-f1
	<-f2
}

func TestWorkerPoolReturnError(t *testing.T) {
	p := NewWorkerPool(5)
	defer p.Close()

	expectedErr := errors.New("HELLO")

	f1 := p.Exec(func() error {
		return expectedErr
	})
	f2 := p.Exec(func() error {
		return nil
	})

	err := <-f1
	<-f2

	assert.Error(t, expectedErr, err)
}

func TestWorkerPoolWithManyTasks(t *testing.T) {
	p := NewWorkerPool(5)
	defer p.Close()

	futures := make([]chan error, 0)
	for i := 0; i < 100; i++ {
		f := p.Exec(func() error {
			fmt.Printf("task %d\n", i)
			return nil
		})
		futures = append(futures, f)
	}

	for _, f := range futures {
		<-f
	}
}
