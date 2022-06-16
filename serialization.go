package running

import (
	"encoding/json"
	"fmt"
)

type JsonPlan struct {
	Props map[string]interface{}

	Parts []Part
}

type Part struct {
	Node *JsonNode

	NextNodes []string
}

type JsonNode struct {
	Name string

	Type string

	SubNodes []*JsonNode

	Wrappers []string
}

func (plan *Plan) MarshalJSON() ([]byte, error) {
	jsonPlan := new(JsonPlan)

	propsMap := make(map[string]interface{})
	if props, ok := plan.Props.(StandardProps); ok {
		propsMap = props
	}
	jsonPlan.Props = propsMap

	if plan.graph == nil {
		if err := plan.Init(); err != nil {
			return nil, err
		}
	}

	for _, vertex := range plan.graph.Vertexes {
		node := newJsonNode(vertex.RefRoot)

		next := make([]string, 0)
		for _, v := range vertex.Next {
			next = append(next, v.RefRoot.NodeName)
		}

		jsonPlan.Parts = append(jsonPlan.Parts, Part{
			Node:      node,
			NextNodes: next,
		})
	}

	return json.Marshal(jsonPlan)
}

func newJsonNode(ref *NodeRef) *JsonNode {
	node := new(JsonNode)
	node.Name = ref.NodeName
	node.Type = ref.NodeType
	node.Wrappers = ref.Wrappers

	for _, subRef := range ref.SubRefs {
		node.SubNodes = append(node.SubNodes, newJsonNode(subRef))
	}

	return node
}

func (plan *Plan) UnmarshalJSON(bytes []byte) error {
	jsonPlan := new(JsonPlan)

	err := json.Unmarshal(bytes, jsonPlan)
	if err != nil {
		return err
	}

	graph := newDAG()
	for _, part := range jsonPlan.Parts {
		parseRefFromJsonNode(part.Node, graph)

		graph.Vertexes[part.Node.Name] = &Vertex{
			RefRoot: graph.NodeRefs[part.Node.Name],
		}
	}

	for _, part := range jsonPlan.Parts {
		if graph.Vertexes[part.Node.Name] == nil {
			continue
		}

		for _, node := range part.NextNodes {
			if graph.Vertexes[node] == nil {
				continue
			}

			graph.Vertexes[part.Node.Name].Next = append(graph.Vertexes[part.Node.Name].Next,
				graph.Vertexes[node])

			graph.Vertexes[node].Prev++
		}
	}

	if err = graph.Verify(); err != nil {
		return fmt.Errorf("invalid plan, %w", err)
	}

	plan.graph = graph
	plan.props = StandardProps(jsonPlan.Props)

	return nil
}

func parseRefFromJsonNode(node *JsonNode, graph *DAG) *NodeRef {
	if graph.NodeRefs[node.Name] != nil {
		return graph.NodeRefs[node.Name]
	}

	ref := &NodeRef{
		NodeName: node.Name,
		NodeType: node.Type,
		Wrappers: node.Wrappers,
	}

	for _, subNode := range node.SubNodes {
		ref.SubRefs = append(ref.SubRefs, parseRefFromJsonNode(subNode, graph))
	}

	graph.NodeRefs[node.Name] = ref
	return ref
}
