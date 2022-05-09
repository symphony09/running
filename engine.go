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
	if engine.pools[name] == nil {
		engine.pools[name] = &WorkerPool{
			sync.Pool{
				New: func() interface{} {
					worker, err := engine.buildWorker(name)
					if err != nil {
						return err
					} else {
						return worker
					}
				},
			},
		}
	}

	output := Output{}
	outputCh := make(chan Output, 1)

	go func() {
		worker, err := engine.pools[name].GetWorker()
		if err != nil {
			output.Err = err
			outputCh <- output
			return
		}
		output = <-worker.Work(ctx)
		outputCh <- output
		engine.pools[name].Put(worker)
	}()

	return outputCh
}

func (engine *Engine) buildWorker(name string) (worker *Worker, err error) {
	if engine.plans[name] == nil {
		err = fmt.Errorf("plan not found, name: %s", name)
		return
	}

	plan := engine.plans[name]
	nodeMap := map[string]Node{}

	if plan.graph == nil {
		plan.graph = newDAG()
		for _, option := range plan.Options {
			option(plan.graph)
		}

		if err = plan.graph.Verify(); err != nil {
			err = fmt.Errorf("invalid plan, %w", err)
			return
		}
	}

	steps, _ := plan.graph.Steps()
	for _, nodeNames := range steps {
		for _, nodeName := range nodeNames {
			if plan.cached && engine.nodeCache[name][nodeName] != nil {
				nodeMap[nodeName] = engine.nodeCache[name][nodeName].(Cloneable).Clone()
			} else {
				nodeMap[nodeName], err = engine.buildNode(plan.graph.NodeRefs[nodeName], plan.Props, "")
				if err != nil {
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

	worker = &Worker{
		steps: steps,
		nodes: nodeMap,
	}
	return
}

func (engine *Engine) buildNode(root *NodeRef, props Props, prefix string) (Node, error) {
	var rootNode Node
	if builder := engine.builders[root.NodeType]; builder != nil {
		if prefix != "" {
			rootNode = builder(prefix+"."+root.NodeName, props)
		} else {
			rootNode = builder(root.NodeName, props)
		}
	} else {
		return nil, fmt.Errorf("no builder found for type %s", root.NodeType)
	}

	//TODO Output warning
	if _, ok := rootNode.(Cluster); !ok {
		return rootNode, nil
	}

	var subNodes []Node
	for _, ref := range root.SubRefs {
		if len(ref.SubRefs) == 0 {
			if builder := engine.builders[ref.NodeType]; builder == nil {
				return nil, fmt.Errorf("no builder found for type %s", ref.NodeType)
			} else {
				subNode := builder(rootNode.Name()+"."+ref.NodeName, props)
				subNodes = append(subNodes, subNode)
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
				subNodes = append(subNodes, subNode)
			}

		}
	}

	rootNode.(Cluster).Inject(subNodes)
	return rootNode, nil
}
