package running

import (
	"context"
	"fmt"
	"sync"
)

type WorkerPool struct {
	sync.Pool
}

func (pool *WorkerPool) GetWorker() (*Worker, error) {
	got := pool.Get()
	if worker, ok := got.(*Worker); ok {
		// return a worker from pool
		return worker, nil
	} else {
		// return error while building worker
		return nil, got.(error)
	}
}

type Worker struct {
	works *WorkList

	nodes map[string]Node

	version string
}

func (worker Worker) Work(ctx context.Context) <-chan Output {
	output := Output{}
	outputCh := make(chan Output, 1)
	state := NewStandardState()

	// get node ready to run from a chan of works, block until all node done
	for nodeName := range worker.works.TODO() {
		go func(nodeName string) {
			defer func() {
				if err := recover(); err != nil {
					output.Err = fmt.Errorf("work panic: %v", err)
				}

				worker.works.Done(nodeName)
			}()

			if statefulNode, ok := worker.nodes[nodeName].(Stateful); ok {
				statefulNode.Bind(state)
			}

			worker.nodes[nodeName].Run(ctx)
			worker.nodes[nodeName].Reset()
		}(nodeName)
	}

	output.State = state
	outputCh <- output
	return outputCh
}

type WorkList struct {
	todo, done chan string

	completed chan struct{}

	Items map[string]*workItem
}

type workItem struct {
	Name string

	State int

	Prev int

	Next []*workItem
}

func NewWorkList(graph *DAG) *WorkList {
	list := &WorkList{
		Items: make(map[string]*workItem),
	}

	for name, vertex := range graph.Vertexes {
		list.Items[name] = &workItem{
			Name:  name,
			State: WorkStateTodo,
			Prev:  vertex.Prev,
			Next:  make([]*workItem, 0),
		}
	}

	for name, vertex := range graph.Vertexes {
		for _, next := range vertex.Next {
			list.Items[name].Next = append(list.Items[name].Next, list.Items[next.RefRoot.NodeName])
		}
	}

	return list
}

func (list *WorkList) TODO() <-chan string {
	list.todo = make(chan string, len(list.Items))
	list.done = make(chan string, len(list.Items))
	list.completed = make(chan struct{}, 1)

	// find node ready to run
	list.feed()

	go func() {
		for {
			select {
			case name := <-list.done:
				if list.Items[name] == nil {
					break
				}

				// mark node done
				list.Items[name].State = WorkStateDone

				for _, nextItem := range list.Items[name].Next {
					nextItem.Prev--
				}

				// find node ready to run
				list.feed()
			case <-list.completed: // all node done, exit
				return
			}
		}
	}()

	return list.todo
}

func (list *WorkList) Done(name string) {
	list.done <- name
}

func (list *WorkList) feed() {
	var hasMoreTodo, hasDoing bool

	// send node ready to run
	for _, item := range list.Items {
		if item.State == WorkStateTodo && item.Prev <= 0 {
			hasMoreTodo = true
			item.State = WorkStateDoing
			list.todo <- item.Name
		}
	}

	// node not found
	if !hasMoreTodo {
		for _, item := range list.Items {
			if item.State == WorkStateDoing {
				hasDoing = true
			}
		}

		// if no nodes are running as well, work is over
		// reset all state,
		// notify goroutine to exits,
		// close chan, end the block.
		if !hasDoing {
			for _, item := range list.Items {
				item.State = WorkStateTodo
				for _, nextItem := range item.Next {
					nextItem.Prev++
				}
			}

			list.completed <- struct{}{}
			close(list.todo)
		}
	}
}
