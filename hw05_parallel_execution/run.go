package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return fmt.Errorf("the number of workers cannot be less than or equal to 0")
	}

	if len(tasks) == 0 {
		return nil
	}

	var errCount int64
	ignoreErrorLimit := m <= 0

	taskCh := make(chan Task, len(tasks))

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(errCount *int64) {
			defer wg.Done()

			for task := range taskCh {
				if atomic.LoadInt64(errCount) >= int64(m) && !ignoreErrorLimit {
					return
				}
				if err := task(); err != nil {
					if ignoreErrorLimit {
						continue
					}
					if atomic.AddInt64(errCount, 1) >= int64(m) {
						return
					}
				}
			}
		}(&errCount)
	}

	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	wg.Wait()

	if errCount >= int64(m) && !ignoreErrorLimit {
		return ErrErrorsLimitExceeded
	}

	return nil
}
