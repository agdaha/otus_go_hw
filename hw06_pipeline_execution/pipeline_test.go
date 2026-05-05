package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

// Граничные условия.
func TestPipelineBoundaryCases(t *testing.T) {
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	// Пайплайн из одного стейджа
	t.Run("single stage", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		stage := g("Stringifier", func(v interface{}) interface{} {
			return strconv.Itoa(v.(int))
		})

		result := make([]string, 0, len(data))
		for s := range ExecutePipeline(in, nil, stage) {
			result = append(result, s.(string))
		}

		require.Equal(t, []string{"1", "2", "3"}, result)
	})

	// Входной канал закрыт сразу
	t.Run("empty input", func(t *testing.T) {
		in := make(Bi)
		close(in)

		stages := []Stage{
			g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
			g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		}

		result := make([]interface{}, 0)
		for v := range ExecutePipeline(in, nil, stages...) {
			result = append(result, v)
		}

		require.Empty(t, result)
	})

	// done закрыт ещё до вызова ExecutePipeline
	t.Run("done closed before start", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		close(done)

		data := []int{1, 2, 3, 4, 5}
		go func() {
			for _, v := range data {
				select {
				case in <- v:
				case <-time.After(sleepPerStage * 2):
				}
			}
			close(in)
		}()

		stages := []Stage{
			g("Dummy", func(v interface{}) interface{} { return v }),
			g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		}

		start := time.Now()
		result := make([]interface{}, 0)
		for v := range ExecutePipeline(in, done, stages...) {
			result = append(result, v)
		}
		elapsed := time.Since(start)

		require.Empty(t, result)
		require.Less(t, int64(elapsed), int64(sleepPerStage)+int64(fault))
	})
}
