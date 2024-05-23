package threading

import (
	"context"
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	numOfExecutions int32
	numOfFailures   int32

	jobError     error
	workersCount chan int
	wg           *sync.WaitGroup
	ctx          context.Context
	cancel       func()
}

func NewWorkerPool(workersCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	if workersCount <= 0 {
		workersCount = 1
	}

	w := &WorkerPool{
		numOfExecutions: int32(0),
		numOfFailures:   int32(0),
		workersCount:    make(chan int, workersCount),
		wg:              &sync.WaitGroup{},
		ctx:             ctx,
		cancel:          cancel,
	}

	for i := 0; i < workersCount; i++ {
		w.workersCount <- 1
	}

	return w
}

func (w *WorkerPool) NumOfExecutions() int32 {
	return atomic.LoadInt32(&w.numOfExecutions)
}

func (w *WorkerPool) NumOfFailures() int32 {
	return atomic.LoadInt32(&w.numOfFailures)
}

func (w *WorkerPool) Error() error {
	return w.jobError
}

func (w *WorkerPool) RunJob(dataset []interface{}, jobFn func(_dataset []interface{}) error) {
	w.wg.Add(<-w.workersCount)

	go func() {
		defer w.wg.Done()
		defer func() { w.workersCount <- 1 }()

		select {
		case <-w.ctx.Done():
			return
		default:
			if err := jobFn(dataset); err != nil {
				atomic.AddInt32(&w.numOfFailures, 1)
				w.jobError = err
			}
		}

		atomic.AddInt32(&w.numOfExecutions, 1)
	}()
}

func (w *WorkerPool) Wait() {
	defer close(w.workersCount)
	w.wg.Wait()
}
