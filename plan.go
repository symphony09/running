package running

import (
	"fmt"
	"strings"
	"sync"
)

type Plan struct {
	Props Props

	Options []Option

	PreparedNodes map[string]Node

	graph *DAG

	version string
}

func NewPlan(props Props, options ...Option) *Plan {
	return &Plan{
		Props: props,

		Options: options,
	}
}

type Option func(*DAG)

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

type NodeRef struct {
	NodeName string

	NodeType string

	SubRefs []*NodeRef
}

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

func (root NodeRef) HasCycle(color map[string]int8) bool {
	color[root.NodeName] = GRAY

	for _, ref := range root.SubRefs {
		if color[ref.NodeName] == GRAY {
			return true
		}

		if color[ref.NodeName] == WHITE && ref.HasCycle(color) {
			return true
		}
	}

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

func (graph *DAG) Steps() ([][]string, []string) {
	steps := make([][]string, 0)

	graph.Lock()
	for {
		var names []string

		for name, vertex := range graph.Vertexes {
			if !vertex.Traversed && vertex.Prev == 0 {
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			break
		}

		for _, name := range names {
			graph.Vertexes[name].Traversed = true
			for _, vertex := range graph.Vertexes[name].Next {
				vertex.Prev--
			}
		}

		steps = append(steps, names)
	}

	left := make([]string, 0)
	for _, vertex := range graph.Vertexes {
		if !vertex.Traversed {
			left = append(left, vertex.RefRoot.NodeName)
		}
	}

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
