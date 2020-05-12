package utils

import (
	"errors"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func TestWorker(t *testing.T) {
	var mutex sync.Mutex

	worker := NewWorker(5)
	indexes := make(map[int]bool)
	for i := 0; i < 30; i++ {
		index := i
		worker.RunTask(func() {
			mutex.Lock()
			indexes[index] = true
			mutex.Unlock()
		})
	}
	worker.Wait()

	for i := 0; i < 30; i++ {
		assert.True(t, indexes[i])
	}

	// support re-add task to worker
	for i := 30; i < 60; i++ {
		index := i
		worker.RunTask(func() {
			mutex.Lock()
			indexes[index] = true
			mutex.Unlock()
		})
	}
	worker.Wait()

	for i := 30; i < 60; i++ {
		assert.True(t, indexes[i])
	}
}

func TestWorkerWaitErr(t *testing.T) {
	worker := NewWorker(5)
	for i := 0; i < 30; i++ {
		worker.RunTask(func() error {
			return errors.New("test")
		})
	}
	err := worker.Wait()
	assert.EqualError(t, err, "test")
	assert.Nil(t, worker.err)
}

func TestWorkerClose(t *testing.T) {
	worker := NewWorker(5)
	for i := 0; i < 30; i++ {
		if i == 8 {
			worker.Close()
		}
		worker.RunTask(func() error {
			return nil
		})
	}
	err := worker.Wait()
	assert.NoError(t, err)
	assert.Nil(t, worker.err)
}
