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

// Cloneable a class of nodes that can be cloned
type Cloneable interface {
	Node

	// Clone self
	Clone() Node
}

// Reversible a class of nodes that can be reverted
type Reversible interface {
	Node

	Revert(ctx context.Context)
}

// Props provide build parameters for the node builder
type Props interface {
	// Get return global value of the key
	Get(key string) (interface{}, bool)

	//SubGet node value of the key, deliver node name as sub
	SubGet(sub, key string) (interface{}, bool)

	// Copy safe use of copies
	Copy() Props
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

type Output struct {
	Err error

	State State
}
