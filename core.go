package running

import (
	"context"
)

type Node interface {
	Run(ctx context.Context)
}

type Cluster interface {
	Node

	Inject(nodes map[string]Node)
}

type Stateful interface {
	Node

	Bind(state State)
}

type Cloneable interface {
	Node

	Clone() Node
}

type Props interface {
	Get(key string) (interface{}, bool)
}

type BuildNodeFunc func(props Props) Node

type State interface {
	Query(key string) (interface{}, bool)

	Update(key string, value interface{})

	Transform(key string, transform TransformStateFunc)
}

type TransformStateFunc func(from interface{}) interface{}

type Base struct {
	State State

	SubNodes map[string]Node
}

func (base *Base) Inject(nodes map[string]Node) {
	base.SubNodes = nodes
}

func (base *Base) Bind(state State) {
	base.State = state

	for _, node := range base.SubNodes {
		if statefulNode, ok := node.(Stateful); ok {
			statefulNode.Bind(state)
		}
	}
}

func (base *Base) Run(ctx context.Context) {
	panic("please implement run method")
}

type Output struct {
	Err error

	State State
}
