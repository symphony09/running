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

	version string

	graph *DAG

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

	plan.version = strconv.FormatInt(time.Now().Unix(), 10)
	plan.graph = graph
	plan.props = plan.Props
	plan.prebuilt = make(map[string]Node)

	for _, node := range plan.Prebuilt {
		plan.prebuilt[node.Name()] = node
	}

	return nil
}

type Option func(*DAG)

// AddNodes add nodes.
// typ declare node type, names declare name of each one.
// node must be added before other options.
var AddNodes = func(typ string, names ...string) Option {
	return func(dag *DAG) {
		for _, name := range names {
			if _, ok := dag.NodeRefs[name]; !ok {
				dag.NodeRefs[name] = &NodeRef{
					NodeName: name,
					NodeType: typ,
				}
			}
		}
	}
}

// MergeNodes merge other nodes as sub-node of the first node.
// example: MergeNodes("A", "B", "C").
// if node "A" implement the Cluster interface, node "B" and "C" will be injected,
// then "A" could use "B" and "C" as sub-nodes.
var MergeNodes = func(cluster string, subNodes ...string) Option {
	return func(dag *DAG) {
		if clusterRef, ok := dag.NodeRefs[cluster]; !ok {
			return
		} else {
			for _, node := range subNodes {
				if _, ok := dag.NodeRefs[node]; ok {
					clusterRef.SubRefs = append(clusterRef.SubRefs, dag.NodeRefs[node])
				}
			}
		}
	}
}

// LinkNodes link first node with others.
// example: LinkNodes("A", "B", "C") => A -> B, A -> C.
var LinkNodes = func(nodes ...string) Option {
	return func(dag *DAG) {
		if len(nodes) < 1 {
			return
		}

		for _, root := range nodes {
			if _, ok := dag.Vertexes[root]; !ok {
				if _, ok := dag.NodeRefs[root]; ok {
					dag.Vertexes[root] = &Vertex{
						RefRoot: dag.NodeRefs[root],
					}
				}
			}
		}

		if dag.Vertexes[nodes[0]] != nil {
			for _, node := range nodes[1:] {
				if dag.Vertexes[node] != nil {
					dag.Vertexes[nodes[0]].Next = append(dag.Vertexes[nodes[0]].Next, dag.Vertexes[node])
					dag.Vertexes[node].Prev++
				}
			}
		}
	}
}

// SLinkNodes link nodes serially.
// example: SLinkNodes("A", "B", "C") => A -> B -> C.
var SLinkNodes = func(nodes ...string) Option {
	return func(dag *DAG) {
		if len(nodes) < 1 {
			return
		}

		for _, root := range nodes {
			if _, ok := dag.Vertexes[root]; !ok {
				if _, ok := dag.NodeRefs[root]; ok {
					dag.Vertexes[root] = &Vertex{
						RefRoot: dag.NodeRefs[root],
					}
				}
			}
		}

		for i := range nodes {
			if i < len(nodes)-1 {
				prev, next := dag.Vertexes[nodes[i]], dag.Vertexes[nodes[i+1]]

				if prev != nil && next != nil {
					prev.Next = append(prev.Next, next)
					next.Prev++
				}
			}
		}
	}
}

type NodeRef struct {
	NodeName string

	NodeType string

	SubRefs []*NodeRef
}

// Print format: NodeA:[SubNodeB, SubNodeC ...]
func (root NodeRef) Print() string {
	var subs []string
	for _, ref := range root.SubRefs {
		if len(ref.SubRefs) == 0 {
			subs = append(subs, ref.NodeName)
		} else {
			subs = append(subs, ref.Print())
		}
	}

	return fmt.Sprintf("%s:[%s]", root.NodeName, strings.Join(subs, ","))
}

const (
	WHITE = 0
	GRAY  = 1
	BLACK = 2
)

// HasCycle detect if there is a circular dependency in a cluster.
func (root NodeRef) HasCycle(color map[string]int8) bool {
	// first visit
	color[root.NodeName] = GRAY

	for _, ref := range root.SubRefs {
		// visit the node twice, there is a circular dependency
		if color[ref.NodeName] == GRAY {
			return true
		}

		// visit sub-node and check cycle
		if color[ref.NodeName] == WHITE && ref.HasCycle(color) {
			return true
		}
	}

	// all sub-nodes are visited
	color[root.NodeName] = BLACK
	return false
}

type Vertex struct {
	Prev int

	Traversed bool

	Next []*Vertex

	RefRoot *NodeRef
}

type DAG struct {
	NodeRefs map[string]*NodeRef

	Vertexes map[string]*Vertex

	sync.Mutex
}

func newDAG() *DAG {
	return &DAG{
		NodeRefs: map[string]*NodeRef{},

		Vertexes: map[string]*Vertex{},
	}
}

// Verify detect if there is a circular dependency.
// include external circular dependency of the graph and internal circular dependency of the cluster.
func (graph *DAG) Verify() error {
	color := make(map[string]int8)

	steps, left := graph.Steps()

	if len(left) != 0 {
		return fmt.Errorf("found cycle between nodes: %v", left)
	}

	for _, nodeNames := range steps {
		for _, nodeName := range nodeNames {
			if graph.NodeRefs[nodeName].HasCycle(color) {
				return fmt.Errorf("found cycle in node: %s", nodeName)
			}
		}
	}

	return nil
}

// Steps list graph steps and left nodes.
// steps example: [[A, B], C], which means that A and B should be processed before C.
// left node can never be processed, implies a circular dependency.
func (graph *DAG) Steps() ([][]string, []string) {
	steps := make([][]string, 0)

	graph.Lock()
	for {
		var names []string

		for name, vertex := range graph.Vertexes {
			// vertex is never traversed and all prev vertex are done, so it can be processed
			if !vertex.Traversed && vertex.Prev == 0 {
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			break
		}

		// set vertex status to traversed
		for _, name := range names {
			graph.Vertexes[name].Traversed = true
			for _, vertex := range graph.Vertexes[name].Next {
				vertex.Prev--
			}
		}

		steps = append(steps, names)
	}

	// found vertexes which are never traversed
	left := make([]string, 0)
	for _, vertex := range graph.Vertexes {
		if !vertex.Traversed {
			left = append(left, vertex.RefRoot.NodeName)
		}
	}

	// reset vertex status to traversed
	graph.reset()
	graph.Unlock()

	return steps, left
}

func (graph *DAG) reset() {
	for _, vertex := range graph.Vertexes {
		vertex.Traversed = false
		for _, nextVertex := range vertex.Next {
			nextVertex.Prev++
		}
	}
}
