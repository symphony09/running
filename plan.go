package running

import (
	"fmt"
	"strings"
)

type Plan struct {
	Props Props

	Options []Option

	graph *DAG

	cached bool
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

type Vertex struct {
	Prev int

	Traversed bool

	Next []*Vertex

	RefRoot *NodeRef
}

type DAG struct {
	NodeRefs map[string]*NodeRef

	Vertexes map[string]*Vertex
}

func newDAG() *DAG {
	return &DAG{
		NodeRefs: map[string]*NodeRef{},

		Vertexes: map[string]*Vertex{},
	}
}

func (graph DAG) Verify() bool {
	//TODO DAG verify
	return true
}

func (graph DAG) Next() []string {
	var names []string

	for name, vertex := range graph.Vertexes {
		if !vertex.Traversed && vertex.Prev == 0 {
			names = append(names, name)
		}
	}

	for _, name := range names {
		graph.Vertexes[name].Traversed = true
		for _, vertex := range graph.Vertexes[name].Next {
			vertex.Prev--
		}
	}

	return names
}

func (graph *DAG) Reset() {
	for _, vertex := range graph.Vertexes {
		vertex.Traversed = false
		for _, nextVertex := range vertex.Next {
			nextVertex.Prev++
		}
	}
}
