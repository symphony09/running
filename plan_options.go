package running

import "fmt"

type Option func(*_DAG)

// AddNodes add nodes.
// typ declare node type, names declare name of each one.
// node must be added before other options.
var AddNodes = func(typ string, names ...string) Option {
	return func(dag *_DAG) {
		for _, name := range names {
			if _, ok := dag.NodeRefs[name]; !ok {
				dag.NodeRefs[name] = &_NodeRef{
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
	return func(dag *_DAG) {
		if clusterRef, ok := dag.NodeRefs[cluster]; !ok {
			dag.Warning = append(dag.Warning, fmt.Sprintf("cluster %s ref not found", cluster))
			return
		} else {
			for _, node := range subNodes {
				if _, ok := dag.NodeRefs[node]; ok {
					clusterRef.SubRefs = append(clusterRef.SubRefs, dag.NodeRefs[node])
				} else {
					dag.Warning = append(dag.Warning, fmt.Sprintf("sub node %s ref not found", node))
				}
			}
		}
	}
}

// WrapNodes wrap node to enhance it,
// wrapper：node type which implement Wrapper,
// targets：wrap targets
var WrapNodes = func(wrapper string, targets ...string) Option {
	return func(dag *_DAG) {
		for _, target := range targets {
			if targetNodeRef := dag.NodeRefs[target]; targetNodeRef != nil {
				targetNodeRef.Wrappers = append(targetNodeRef.Wrappers, wrapper)
			} else {
				dag.Warning = append(dag.Warning, fmt.Sprintf("wrap target node %s ref not found", target))
			}
		}
	}
}

// WrapAllNodes wrap all nodes with single or multi wrappers,
// will only affect nodes added before this
var WrapAllNodes = func(wrappers ...string) Option {
	return func(dag *_DAG) {
		for _, wrapper := range wrappers {
			for _, ref := range dag.NodeRefs {
				if ref != nil {
					ref.Wrappers = append(ref.Wrappers, wrapper)
				}
			}
		}
	}
}

// ReUseNodes reuse node to avoid unnecessary rebuilds,
// fits nodes whose properties do not change and implements the clone method
var ReUseNodes = func(nodes ...string) Option {
	return func(dag *_DAG) {
		for _, node := range nodes {
			if dag.NodeRefs[node] != nil {
				dag.NodeRefs[node].ReUse = true
			}
		}
	}
}

// MarkNodes set label of nodes
var MarkNodes = func(label string, nodes ...string) Option {
	return func(dag *_DAG) {
		for _, node := range nodes {
			if dag.NodeRefs[node] != nil {
				ref := dag.NodeRefs[node]
				if ref.Labels == nil {
					ref.Labels = make(map[string]struct{})
				}

				ref.Labels[label] = struct{}{}
			}
		}
	}
}

// LinkNodes link first node with others.
// example: LinkNodes("A", "B", "C") => A -> B, A -> C.
var LinkNodes = func(nodes ...string) Option {
	return func(dag *_DAG) {
		if len(nodes) < 1 {
			return
		}

		for _, root := range nodes {
			if _, ok := dag.Vertexes[root]; !ok {
				if _, ok := dag.NodeRefs[root]; ok {
					dag.Vertexes[root] = &_Vertex{
						RefRoot: dag.NodeRefs[root],
					}
				} else {
					dag.Warning = append(dag.Warning, fmt.Sprintf("link target node %s ref not found", root))
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
	return func(dag *_DAG) {
		if len(nodes) < 1 {
			return
		}

		for _, root := range nodes {
			if _, ok := dag.Vertexes[root]; !ok {
				if _, ok := dag.NodeRefs[root]; ok {
					dag.Vertexes[root] = &_Vertex{
						RefRoot: dag.NodeRefs[root],
					}
				} else {
					dag.Warning = append(dag.Warning, fmt.Sprintf("Slink target node %s ref not found", root))
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

// RLinkNodes link first node with others.
// example: RLinkNodes("A", "B", "C") => B -> A, C -> A.
var RLinkNodes = func(nodes ...string) Option {
	return func(dag *_DAG) {
		if len(nodes) < 1 {
			return
		}

		for _, root := range nodes {
			if _, ok := dag.Vertexes[root]; !ok {
				if _, ok := dag.NodeRefs[root]; ok {
					dag.Vertexes[root] = &_Vertex{
						RefRoot: dag.NodeRefs[root],
					}
				} else {
					dag.Warning = append(dag.Warning, fmt.Sprintf("link target node %s ref not found", root))
				}
			}
		}

		if dag.Vertexes[nodes[0]] != nil {
			for _, node := range nodes[1:] {
				if dag.Vertexes[node] != nil {
					dag.Vertexes[node].Next = append(dag.Vertexes[node].Next, dag.Vertexes[nodes[0]])
					dag.Vertexes[nodes[0]].Prev++
				}
			}
		}
	}
}
