package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func runWorkers(cancel chan struct{}, taskCh chan Task, wg *sync.WaitGroup, n int, limit *int64, unlim bool) {
	var isCancelled int32 // статус канала cancel 0 = в работе, 1 = остановлен

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-cancel:
					return
				default:
					err := task()
					if err != nil {
						atomic.AddInt64(limit, -1)
					}
					if atomic.LoadInt64(limit) <= 0 && !unlim {
						if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {
							close(cancel)
						}
						return
					}
				}
			}
		}()
	}
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return fmt.Errorf("the number of workers cannot be less than or equal to 0")
	}

	if len(tasks) == 0 {
		return nil
	}

	errLimit := int64(m)
	ignoreErrorLimit := m <= 0 // игнорировать лимит ошибок если он <= 0

	taskCh := make(chan Task)
	cancel := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(n)

	runWorkers(cancel, taskCh, &wg, n, &errLimit, ignoreErrorLimit)

pusher:
	for _, task := range tasks {
		select {
		case <-cancel:
			break pusher
		case taskCh <- task:
		}
	}
	close(taskCh)

	wg.Wait()

	if errLimit <= 0 && !ignoreErrorLimit {
		return ErrErrorsLimitExceeded
	}

	return nil
}
