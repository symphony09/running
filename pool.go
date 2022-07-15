package running

import (
	"sync"
)

type _WorkerPool struct {
	sync.Pool
}

func (pool *_WorkerPool) GetWorker() (*_Worker, error) {
	got := pool.Get()
	if worker, ok := got.(*_Worker); ok {
		// return a worker from pool
		return worker, nil
	} else {
		// return error while building worker
		return nil, got.(error)
	}
}
