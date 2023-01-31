package running

import (
	"context"
	"fmt"
	"sync"
)

const (
	_WorkStatusTodo  = 0
	_WorkStatusDoing = 1
	_WorkStatusDone  = 2
)

type _Worker struct {
	Works *_WorkList

	Nodes map[string]Node

	StateBuilder func() State

	Version string
}

func (worker _Worker) Work(ctx context.Context) <-chan Output {
	output := Output{}
	outputCh := make(chan Output, 1)
	state := worker.StateBuilder()

	skipNodes := make(map[string]struct{})
	raw := ctx.Value(CtxKey)
	if raw != nil {
		if params, ok := raw.(CtxParams); ok {
			for _, node := range params.SkipNodes {
				skipNodes[node] = struct{}{}
			}
		}
	}

	// get node ready to run from a chan of works, block until all node done
	for nodeName := range worker.Works.TODO() {
		go func(nodeName string) {
			if _, ok := skipNodes[nodeName]; ok {
				worker.Works.Done(nodeName)
				return
			}

			defer func() {
				if err := recover(); err != nil {
					output.Err = fmt.Errorf("%w, node name: %s, panic info: %v", ErrWorkerPanic, nodeName, err)
					worker.Works.Terminate(nodeName)
				} else {
					worker.Works.Done(nodeName)
				}
			}()

			if statefulNode, ok := worker.Nodes[nodeName].(Stateful); ok {
				statefulNode.Bind(state)
			}

			worker.Nodes[nodeName].Run(ctx)
			worker.Nodes[nodeName].Reset()
		}(nodeName)
	}

	output.State = state
	outputCh <- output
	return outputCh
}

type _WorkList struct {
	todo, done chan string

	completed chan struct{}

	terminate chan string

	Items map[string]*_WorkItem

	sync.RWMutex
}

type _WorkItem struct {
	Name string

	Status int

	Prev int

	Next []*_WorkItem
}

func newWorkList(graph *_DAG) *_WorkList {
	list := &_WorkList{
		Items: make(map[string]*_WorkItem),
	}

	for name, vertex := range graph.Vertexes {
		list.Items[name] = &_WorkItem{
			Name:   name,
			Status: _WorkStatusTodo,
			Prev:   vertex.Prev,
			Next:   make([]*_WorkItem, 0),
		}
	}

	for name, vertex := range graph.Vertexes {
		for _, next := range vertex.Next {
			list.Items[name].Next = append(list.Items[name].Next, list.Items[next.RefRoot.NodeName])
		}
	}

	return list
}

func (list *_WorkList) TODO() <-chan string {
	list.Lock()

	list.todo = make(chan string, len(list.Items))
	list.done = make(chan string, len(list.Items))
	list.completed = make(chan struct{}, 1)
	list.terminate = make(chan string, len(list.Items))

	list.Unlock()

	// find node ready to run
	list.feed()

	go func() {
		list.RLock()
		defer list.RUnlock()

		for {
			select {
			case name := <-list.done:
				if list.Items[name] == nil {
					break
				}

				// mark node done
				list.Items[name].Status = _WorkStatusDone

				for _, nextItem := range list.Items[name].Next {
					nextItem.Prev--
				}

				// find node ready to run
				list.feed()
			case name := <-list.terminate:
				if list.Items[name] == nil {
					break
				}

				// mark node done
				list.Items[name].Status = _WorkStatusDone

				// no more nodes need to do
				for _, item := range list.Items {
					if item.Status == _WorkStatusTodo {
						item.Status = _WorkStatusDone
					}
				}

				// can't return here, wait all node done
				list.feed()
			case <-list.completed: // all node done, exit
				return
			}
		}
	}()

	return list.todo
}

func (list *_WorkList) Terminate(name string) {
	list.terminate <- name
}

func (list *_WorkList) Done(name string) {
	list.done <- name
}

// notify goroutine to exits,
// close chan, end the block.
func (list *_WorkList) clean() {
	list.completed <- struct{}{}

	for _, item := range list.Items {
		item.Status = _WorkStatusTodo
		item.Prev = 0
	}

	for _, item := range list.Items {
		for _, nextItem := range item.Next {
			nextItem.Prev++
		}
	}

	close(list.todo)
}

func (list *_WorkList) feed() {
	var hasMoreTodo, hasDoing bool

	// send node ready to run
	for _, item := range list.Items {
		if item.Status == _WorkStatusTodo && item.Prev <= 0 {
			hasMoreTodo = true
			item.Status = _WorkStatusDoing
			list.todo <- item.Name
		}
	}

	// node not found
	if !hasMoreTodo {
		for _, item := range list.Items {
			if item.Status == _WorkStatusDoing {
				hasDoing = true
			}
		}

		// if no nodes are running as well, work is over
		if !hasDoing {
			list.clean()
		}
	}
}
