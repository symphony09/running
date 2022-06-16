package running

import (
	"context"
)

// Node basic unit of execution
type Node interface {
	Name() string

	// Run will be called when all deps solved or cluster invoke it
	Run(ctx context.Context)

	// Reset will be called when the node will no longer execute until the next execution plan
	Reset()
}

// Cluster a class of nodes that can contain other nodes
type Cluster interface {
	Node

	// Inject deliver the sub-nodes, will be called when engine build the cluster
	Inject(nodes []Node)
}

// Wrapper a class of nodes that can wrap other node
type Wrapper interface {
	Node

	Wrap(target Node)
}

// Stateful a class of nodes that need record or query state
type Stateful interface {
	Node

	// Bind deliver the state, should be called before engine run the node
	Bind(state State)
}

//Cloneable a class of nodes that can be cloned
type Cloneable interface {
	Node

	// Clone self
	Clone() Node
}

// Props provide build parameters for the node builder
type Props interface {
	// Get return global value of the key
	Get(key string) (interface{}, bool)

	//SubGet node value of the key, deliver node name as sub
	SubGet(sub, key string) (interface{}, bool)
}

type BuildNodeFunc func(name string, props Props) (Node, error)

// State store state of nodes
type State interface {
	// Query return value of the key
	Query(key string) (interface{}, bool)

	// Update set a new value for the key
	Update(key string, value interface{})

	// Transform set a new value for the key, according to the old value
	Transform(key string, transform TransformStateFunc)
}

type TransformStateFunc func(from interface{}) interface{}

// Base a simple impl of Node, Cluster, Stateful
// Embed it in custom node and override interface methods as needed
type Base struct {
	NodeName string

	State State

	SubNodes []Node

	SubNodesMap map[string]Node
}

func (base *Base) SetName(name string) {
	base.NodeName = name
}

func (base *Base) Name() string {
	return base.NodeName
}

func (base *Base) Inject(nodes []Node) {
	base.SubNodes = append(base.SubNodes, nodes...)

	if base.SubNodesMap == nil {
		base.SubNodesMap = make(map[string]Node)
	}

	for _, node := range nodes {
		base.SubNodesMap[node.Name()] = node
	}
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

type BaseWrapper struct {
	Target Node
}

func (wrapper *BaseWrapper) Wrap(target Node) {
	wrapper.Target = target
}

func (wrapper *BaseWrapper) Name() string {
	return wrapper.Target.Name()
}

func (wrapper *BaseWrapper) Run(ctx context.Context) {
	wrapper.Target.Run(ctx)
}

func (wrapper *BaseWrapper) Reset() {
	wrapper.Target.Reset()
}

func (wrapper *BaseWrapper) Bind(state State) {
	if statefulTarget, ok := wrapper.Target.(Stateful); ok {
		statefulTarget.Bind(state)
	}
}

type Output struct {
	Err error

	State State
}
