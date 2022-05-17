package running

import (
	"context"
	"fmt"
	"sync"
)

var Global = &Engine{
	builders: map[string]BuildNodeFunc{},

	plans: map[string]*Plan{},

	pools: map[string]*WorkerPool{},
}

type Engine struct {
	builders map[string]BuildNodeFunc

	plans map[string]*Plan

	pools map[string]*WorkerPool
}

func (engine *Engine) RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	engine.builders[name] = builder
}

func (engine *Engine) RegisterPlan(name string, plan *Plan) error {
	err := plan.Init()
	if err != nil {
		return err
	}
	engine.plans[name] = plan
	return nil
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

		worker, err := engine.pools[name].GetWorker()
		if err != nil {
			output.Err = err
			outputCh <- output
			return
		}
		output = <-worker.Work(ctx)
		outputCh <- output

		if worker.version == engine.plans[name].version {
			engine.pools[name].Put(worker)
		}
	}()

	return outputCh
}

func (engine *Engine) UpdatePlan(name string, fastMode bool, update func(plan *Plan)) error {
	plan := engine.plans[name]

	plan.locker.Lock()
	update(plan)
	plan.locker.Unlock()

	err := plan.Init()
	if err != nil {
		return err
	}

	if fastMode {
		engine.pools[name] = nil
	}

	return nil
}

func (engine *Engine) buildWorker(name string) (worker *Worker, err error) {
	plan := engine.plans[name]

	plan.locker.RLock()
	defer plan.locker.RUnlock()

	nodeMap := map[string]Node{}

	prebuilt := plan.prebuilt

	steps, _ := plan.graph.Steps()
	for _, nodeNames := range steps {
		for _, nodeName := range nodeNames {
			if prebuilt[nodeName] != nil {
				if cloneableNode, ok := prebuilt[nodeName].(Cloneable); ok {
					nodeMap[nodeName] = cloneableNode.Clone()
				} else {
					nodeMap[nodeName] = prebuilt[nodeName]
				}
			} else {
				nodeMap[nodeName], err = engine.buildNode(plan.graph.NodeRefs[nodeName], plan.props, "")
				if err != nil {
					return
				}
			}
		}
	}

	worker = &Worker{
		steps:   steps,
		nodes:   nodeMap,
		version: plan.version,
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
