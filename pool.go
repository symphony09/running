package running

import (
	"context"
	"sync"
)

type WorkerPool struct {
	sync.Pool
}

func (pool *WorkerPool) GetWorker() Worker {
	return pool.Get().(Worker)
}

type Worker struct {
	steps [][]string

	nodes map[string]Node
}

func (worker Worker) Work(ctx context.Context) <-chan Output {
	output := Output{}
	outputCh := make(chan Output, 1)
	state := NewStandardState()

	var wg sync.WaitGroup

	for _, nodeNames := range worker.steps {
		for _, nodeName := range nodeNames {
			wg.Add(1)

			if statefulNode, ok := worker.nodes[nodeName].(Stateful); ok {
				statefulNode.Bind(state)
			}

			go func(nodeName string) {
				worker.nodes[nodeName].Run(ctx)

				wg.Done()
			}(nodeName)
		}

		wg.Wait()
	}

	output.State = state
	outputCh <- output
	return outputCh
}
