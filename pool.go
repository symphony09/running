package running

import (
	"sync"
)

type _WorkerPool struct {
	sync.Pool

	bufQueue []*_Worker
	qFront   int
	qRear    int
	mu       sync.Mutex
}

func (pool *_WorkerPool) GetWorker() (*_Worker, error) {
	if len(pool.bufQueue) > 0 {
		if worker := pool.qGet(); worker != nil {
			return worker, nil
		}
	}

	got := pool.Get()
	if worker, ok := got.(*_Worker); ok {
		// return a worker from pool
		return worker, nil
	} else {
		// return error while building worker
		return nil, got.(error)
	}
}

func (pool *_WorkerPool) PutWorker(worker *_Worker) {
	if len(pool.bufQueue) > 0 && pool.qPut(worker) {
		return
	}

	pool.Put(worker)
}

func (pool *_WorkerPool) Warmup(size int) {
	if size <= 0 {
		return
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.bufQueue = make([]*_Worker, size)
	for i := 0; i < size; i++ {
		got := pool.Get()
		if worker, ok := got.(*_Worker); ok {
			pool.bufQueue[i] = worker
		}
	}

	pool.qFront = 0
	pool.qRear = size - 1
}

func (pool *_WorkerPool) qGet() (worker *_Worker) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.qFront == pool.qRear {
		return
	}

	worker = pool.bufQueue[pool.qFront]
	pool.bufQueue[pool.qFront] = nil
	pool.qFront = (pool.qFront + 1) % len(pool.bufQueue)
	return
}

func (pool *_WorkerPool) qPut(worker *_Worker) (success bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if (pool.qRear+1)%len(pool.bufQueue) == pool.qFront {
		return
	}

	pool.bufQueue[pool.qRear] = worker
	pool.qRear = (pool.qRear + 1) % len(pool.bufQueue)
	return true
}
