package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in
	for _, stage := range stages {
		out = stageRunner(done, out, stage)
	}
	return out
}

func stageRunner(done In, in In, stage Stage) Out {
	stageIn := make(Bi)

	go func() {
		defer close(stageIn)
		for {
			select {
			case <-done:
				go Drain(in)
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					go Drain(in)
					return
				case stageIn <- v:
				}
			}
		}
	}()

	return stage(stageIn)
}

// Функция Drain опустошает входной канал с целью исключение блокировки
// если предыдущий этап оказался на этапе записи в канал.
func Drain(in In) {
	for range in { //nolint:revive
	}
}
