package running

import (
	"encoding/json"
	"fmt"
)

type JsonPlan struct {
	Props json.RawMessage

	Graph []GraphNode
}

type GraphNode struct {
	Node *JsonNode

	NextNodes []string
}

type JsonNode struct {
	Name string

	Type string

	SubNodes []*JsonNode

	Wrappers []string

	ReUse bool
}

func (plan *Plan) MarshalJSON() ([]byte, error) {
	jsonPlan := new(JsonPlan)

	if plan.graph == nil {
		if err := plan.Init(); err != nil {
			return nil, err
		}
	}

	if exportable, ok := plan.props.(ExportableProps); ok {
		propsData, err := json.Marshal(exportable.Raw())
		if err == nil {
			jsonPlan.Props = propsData
		} else {
			return nil, err
		}
	}

	for _, vertex := range plan.graph.Vertexes {
		node := newJsonNode(vertex.RefRoot)

		next := make([]string, 0)
		for _, v := range vertex.Next {
			next = append(next, v.RefRoot.NodeName)
		}

		jsonPlan.Graph = append(jsonPlan.Graph, GraphNode{
			Node:      node,
			NextNodes: next,
		})
	}

	return json.Marshal(jsonPlan)
}

func newJsonNode(ref *_NodeRef) *JsonNode {
	node := new(JsonNode)
	node.Name = ref.NodeName
	node.Type = ref.NodeType
	node.Wrappers = ref.Wrappers
	node.ReUse = ref.ReUse

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
	for _, part := range jsonPlan.Graph {
		parseRefFromJsonNode(part.Node, graph)

		graph.Vertexes[part.Node.Name] = &_Vertex{
			RefRoot: graph.NodeRefs[part.Node.Name],
		}
	}

	for _, part := range jsonPlan.Graph {
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

	propsMap := make(map[string]interface{})
	err = json.Unmarshal(jsonPlan.Props, &propsMap)
	if err != nil {
		return err
	} else {
		plan.props = StandardProps(propsMap)
		plan.Props = plan.props.Copy()
	}

	return nil
}

func parseRefFromJsonNode(node *JsonNode, graph *_DAG) *_NodeRef {
	if graph.NodeRefs[node.Name] != nil {
		return graph.NodeRefs[node.Name]
	}

	ref := &_NodeRef{
		NodeName: node.Name,
		NodeType: node.Type,
		Wrappers: node.Wrappers,
		ReUse:    node.ReUse,
	}

	for _, subNode := range node.SubNodes {
		ref.SubRefs = append(ref.SubRefs, parseRefFromJsonNode(subNode, graph))
	}

	graph.NodeRefs[node.Name] = ref
	return ref
}
