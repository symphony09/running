package running

import (
	"fmt"
	"strings"
	"sync"
)

type _DAG struct {
	NodeRefs map[string]*_NodeRef

	Vertexes map[string]*_Vertex

	Warning []string

	sync.Mutex
}

type _Vertex struct {
	Prev int

	Traversed bool

	Next []*_Vertex

	RefRoot *_NodeRef
}

func newDAG() *_DAG {
	return &_DAG{
		NodeRefs: map[string]*_NodeRef{},

		Vertexes: map[string]*_Vertex{},
	}
}

// Verify detect if there is a circular dependency.
// include external circular dependency of the graph and internal circular dependency of the cluster.
func (graph *_DAG) Verify() error {
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
func (graph *_DAG) Steps() ([][]string, []string) {
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

func (graph *_DAG) reset() {
	for _, vertex := range graph.Vertexes {
		vertex.Traversed = false
		for _, nextVertex := range vertex.Next {
			nextVertex.Prev++
		}
	}
}

type _NodeRef struct {
	NodeName string

	NodeType string

	SubRefs []*_NodeRef

	Wrappers []string

	ReUse bool
}

const (
	_WHITE = 0
	_GRAY  = 1
	_BLACK = 2
)

// HasCycle detect if there is a circular dependency in a cluster.
func (root _NodeRef) HasCycle(color map[string]int8) bool {
	// first visit
	color[root.NodeName] = _GRAY

	for _, ref := range root.SubRefs {
		// visit the node twice, there is a circular dependency
		if color[ref.NodeName] == _GRAY {
			return true
		}

		// visit sub-node and check cycle
		if color[ref.NodeName] == _WHITE && ref.HasCycle(color) {
			return true
		}
	}

	// all sub-nodes are visited
	color[root.NodeName] = _BLACK
	return false
}

// Print format: NodeA:[SubNodeB, SubNodeC ...]
func (root _NodeRef) Print() string {
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
