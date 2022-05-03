package running

import (
	"context"
	"fmt"
)

var Global = &Engine{
	builders: map[string]BuildNodeFunc{},

	plans: map[string]*Plan{},

	nodeCache: map[string]map[string]Node{},
}

type Engine struct {
	builders map[string]BuildNodeFunc

	plans map[string]*Plan

	nodeCache map[string]map[string]Node
}

func (engine *Engine) RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	engine.builders[name] = builder
}

func (engine *Engine) RegisterPlan(name string, plan *Plan) {
	engine.plans[name] = plan
	engine.nodeCache[name] = map[string]Node{}
}

func (engine *Engine) ExecPlan(name string, ctx context.Context) <-chan Output {
	output := Output{}
	outputCh := make(chan Output, 1)

	go func() {
		if engine.plans[name] == nil {
			output.Err = fmt.Errorf("plan not found, name: %s", name)
			outputCh <- output
			return
		}

		plan := engine.plans[name]
		state := NewStandardState()
		nodeMap := map[string]Node{}

		if plan.graph == nil {
			plan.graph = newDAG()
			for _, option := range plan.Options {
				option(plan.graph)
			}
		}

		nodeNames := plan.graph.Next()
		for len(nodeNames) > 0 {
			for _, nodeName := range nodeNames {
				if plan.cached && engine.nodeCache[name][nodeName] != nil {
					nodeMap[nodeName] = engine.nodeCache[name][nodeName].(Cloneable).Clone()
				} else {
					nodeMap[nodeName], output.Err = engine.buildNode(plan.graph.NodeRefs[nodeName], plan.Props)
					if output.Err != nil {
						outputCh <- output
						return
					}
				}

				if statefulNode, ok := nodeMap[nodeName].(Stateful); ok {
					statefulNode.Bind(state)
				}
			}

			nodeNames = plan.graph.Next()
		}

		if !plan.cached {
			for nodeName, node := range nodeMap {
				if _, ok := node.(Cloneable); ok {
					engine.nodeCache[name][nodeName] = node
				}
			}

			plan.cached = true
		}

		plan.graph.Reset()
		nodeNames = plan.graph.Next()
		for len(nodeNames) > 0 {
			for _, nodeName := range nodeNames {
				nodeMap[nodeName].Run(ctx)
			}

			nodeNames = plan.graph.Next()
		}

		output.State = state
		outputCh <- output
		return
	}()

	return outputCh
}

func (engine *Engine) buildNode(root *NodeRef, props Props) (Node, error) {
	var rootNode Node
	if builder := engine.builders[root.NodeType]; builder != nil {
		rootNode = builder(props)
	} else {
		return nil, fmt.Errorf("no builder found for type %s", root.NodeType)
	}

	//TODO Output warning
	if _, ok := rootNode.(Cluster); !ok {
		return rootNode, nil
	}

	subNodes := map[string]Node{}
	for _, ref := range root.SubRefs {
		if len(ref.SubRefs) == 0 {
			if builder := engine.builders[ref.NodeType]; builder == nil {
				return nil, fmt.Errorf("no builder found for type %s", ref.NodeType)
			} else {
				subNodes[ref.NodeName] = builder(props)
			}
		} else {
			if subNode, err := engine.buildNode(ref, props); err != nil {
				return nil, err
			} else {
				subNodes[ref.NodeName] = subNode
			}

		}
	}

	rootNode.(Cluster).Inject(subNodes)
	return rootNode, nil
}
