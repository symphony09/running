package running

import (
	"context"
)

type Node interface {
	Name() string

	Run(ctx context.Context)

	Reset()
}

type Cluster interface {
	Node

	Inject(nodes []Node)
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

	SubGet(sub, key string) (interface{}, bool)
}

type BuildNodeFunc func(name string, props Props) (Node, error)

type State interface {
	Query(key string) (interface{}, bool)

	Update(key string, value interface{})

	Transform(key string, transform TransformStateFunc)
}

type TransformStateFunc func(from interface{}) interface{}

type Base struct {
	NodeName string

	State State

	SubNodes []Node
}

func (base *Base) SetName(name string) {
	base.NodeName = name
}

func (base *Base) Name() string {
	return base.NodeName
}

func (base *Base) Inject(nodes []Node) {
	base.SubNodes = append(base.SubNodes, nodes...)
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

func (base *Base) Reset() {
	base.ResetSubNodes()
}

func (base *Base) ResetSubNodes() {
	for _, node := range base.SubNodes {
		node.Reset()
	}
}

type Output struct {
	Err error

	State State
}
