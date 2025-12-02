package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("tasks ulimit count errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 10
		maxErrorsCount := 0

		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})
}

func TestRun_HomeWork(t *testing.T) {
	t.Run("tasks with Eventually", func(t *testing.T) {
		workersCount := 5
		tasksCount := 20

		var activeTasks int32

		tasks := make([]Task, tasksCount)
		for i := 0; i < tasksCount; i++ {
			tasks[i] = func() error {
				atomic.AddInt32(&activeTasks, 1)
				defer atomic.AddInt32(&activeTasks, -1)
				time.Sleep(100 * time.Millisecond)
				return nil
			}
		}

		doneCh := make(chan error)
		go func() {
			doneCh <- Run(tasks, workersCount, 0)
		}()

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&activeTasks) == int32(workersCount)
		}, 1*time.Second, 10*time.Millisecond, "expected %d workers to run concurrently", workersCount)

		err := <-doneCh
		require.NoError(t, err)
	})

	t.Run("stress test (go test -race)", func(t *testing.T) {
		tasksCount := 10000
		workersCount := 500
		maxErrorsCount := 3000

		tasks := make([]Task, tasksCount)
		for i := 0; i < tasksCount; i++ {
			if i%3 != 0 {
				tasks[i] = func() error { return nil }
			} else {
				tasks[i] = func() error { return fmt.Errorf("Ups!!!") }
			}
		}

		err := Run(tasks, workersCount, maxErrorsCount)
		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
	})
}
