package running

import (
	"context"
	"fmt"
	"sync"
)

var Global = &Engine{
	builders: map[string]BuildNodeFunc{},

	plans: map[string]*Plan{},

	nodeCache: map[string]map[string]Node{},

	pools: map[string]*WorkerPool{},
}

type Engine struct {
	builders map[string]BuildNodeFunc

	plans map[string]*Plan

	nodeCache map[string]map[string]Node

	pools map[string]*WorkerPool
}

func (engine *Engine) RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	engine.builders[name] = builder
}

func (engine *Engine) RegisterPlan(name string, plan *Plan) {
	engine.plans[name] = plan
	engine.nodeCache[name] = map[string]Node{}
}

func (engine *Engine) ExecPlan(name string, ctx context.Context) <-chan Output {
	if engine.pools[name] != nil {
		return engine.pools[name].GetWorker().Work(ctx)
	}

	output := Output{}
	outputCh := make(chan Output, 1)

	go func() {
		if engine.plans[name] == nil {
			output.Err = fmt.Errorf("plan not found, name: %s", name)
			outputCh <- output
			return
		}

		plan := engine.plans[name]
		nodeMap := map[string]Node{}

		if plan.graph == nil {
			plan.graph = newDAG()
			for _, option := range plan.Options {
				option(plan.graph)
			}

			if err := plan.graph.Verify(); err != nil {
				output.Err = fmt.Errorf("invalid plan, %w", err)
				outputCh <- output
				return
			}
		}

		steps, _ := plan.graph.Steps()
		for _, nodeNames := range steps {
			for _, nodeName := range nodeNames {
				if plan.cached && engine.nodeCache[name][nodeName] != nil {
					nodeMap[nodeName] = engine.nodeCache[name][nodeName].(Cloneable).Clone()
				} else {
					nodeMap[nodeName], output.Err = engine.buildNode(plan.graph.NodeRefs[nodeName], plan.Props, "")
					if output.Err != nil {
						outputCh <- output
						return
					}
				}
			}
		}

		if !plan.cached {
			for nodeName, node := range nodeMap {
				if _, ok := node.(Cloneable); ok {
					engine.nodeCache[name][nodeName] = node
				}
			}

			plan.cached = true
		}

		pool := &WorkerPool{
			sync.Pool{
				New: func() interface{} {
					return Worker{
						steps: steps,
						nodes: nodeMap,
					}
				},
			},
		}

		worker := pool.GetWorker()
		output = <-worker.Work(ctx)
		outputCh <- output
		pool.Put(worker)

		engine.pools[name] = pool
		return
	}()

	return outputCh
}

func (engine *Engine) buildNode(root *NodeRef, props Props, prefix string) (Node, error) {
	var rootNode Node
	if builder := engine.builders[root.NodeType]; builder != nil {
		rootNode = builder(props)
		if prefix != "" {
			rootNode.SetName(prefix + "." + root.NodeName)
		} else {
			rootNode.SetName(root.NodeName)
		}
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
				subNodes[ref.NodeName].SetName(rootNode.Name() + "." + ref.NodeName)
			}
		} else {
			if prefix == "" {
				prefix = root.NodeName
			} else {
				prefix = prefix + "." + root.NodeName
			}
			if subNode, err := engine.buildNode(ref, props, prefix); err != nil {
				return nil, err
			} else {
				subNodes[ref.NodeName] = subNode
			}

		}
	}

	rootNode.(Cluster).Inject(subNodes)
	return rootNode, nil
}
