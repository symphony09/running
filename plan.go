package running

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Plan explain how to execute nodes
type Plan struct {
	Props Props

	Prebuilt []Node

	Options []Option

	Strict bool

	version string

	graph *_DAG

	props Props

	prebuilt map[string]Node

	locker sync.RWMutex
}

// NewPlan new a plan.
// props: build props of nodes.
// prebuilt: prebuilt nodes, reduce cost of build node, nil is fine.
// options: AddNodes, LinkNodes and so on.
func NewPlan(props Props, prebuilt []Node, options ...Option) *Plan {
	return &Plan{
		Props: props,

		Prebuilt: prebuilt,

		Options: options,
	}
}

// Init Plan take effect only after initialization.
// if plan is invalid, such as circular dependencies, return error.
func (plan *Plan) Init() error {
	plan.locker.Lock()
	defer plan.locker.Unlock()

	graph := newDAG()
	for _, option := range plan.Options {
		option(graph)
	}
	if err := graph.Verify(); err != nil {
		return fmt.Errorf("invalid plan, %w", err)
	}

	if plan.Strict && len(graph.Warning) > 0 {
		return fmt.Errorf("invaild plan, %s", strings.Join(graph.Warning, ";"))
	}

	plan.version = strconv.FormatInt(time.Now().Unix(), 10)
	plan.graph = graph
	plan.props = plan.Props
	plan.prebuilt = make(map[string]Node)

	for _, node := range plan.Prebuilt {
		plan.prebuilt[node.Name()] = node
	}

	return nil
}
