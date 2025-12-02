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

	ignoreErrorLimit := m <= 0 // игнорировать лимит ошибок если он <= 0
	errLimit := int64(m)       // количество ошибок, которые можно пропустить, приведено в int64 для атомикова

	taskCh := make(chan Task)
	cancel := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(n)

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
						atomic.AddInt64(&errLimit, -1)
					}
					if atomic.LoadInt64(&errLimit) <= 0 && !ignoreErrorLimit {
						if atomic.CompareAndSwapInt32(&isCancelled, 0, 1) {
							close(cancel)
						}
						return
					}
				}
			}
		}()
	}

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
