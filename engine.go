package running

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var Global = NewDefaultEngine()

func NewDefaultEngine() *Engine {
	return &Engine{
		StateBuilder: func() State {
			return NewStandardState()
		},

		builders: map[string]BuildNodeFunc{},

		plans: map[string]*Plan{},

		pools: map[string]*_WorkerPool{},
	}
}

// RegisterNodeBuilder register node builder to Global
func RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	Global.RegisterNodeBuilder(name, builder)
}

// RegisterPlan register plan to Global
func RegisterPlan(name string, plan *Plan) error {
	return Global.RegisterPlan(name, plan)
}

// ExecPlan exec plan register in Global
func ExecPlan(name string, ctx context.Context) <-chan Output {
	return Global.ExecPlan(name, ctx)
}

// UpdatePlan update plan register in Global.
func UpdatePlan(name string, update func(plan *Plan)) error {
	return Global.UpdatePlan(name, update)
}

// ExportPlan export plan register in Global, return json bytes
func ExportPlan(name string) ([]byte, error) {
	return Global.ExportPlan(name)
}

// WarmupPool warm up pool to avoid cold start
// name: plan name
// size: set size of worker buf queue
func WarmupPool(name string, size int) {
	Global.WarmupPool(name, size)
}

// ClearPool clear worker pool of plan, invoke it to make plan effect immediately after update
// name: name of plan
func ClearPool(name string) {
	Global.ClearPool(name)
}

// LoadPlanFromJson load plan from json data
// name: name of plan to load
// jsonData: json data of plan
// prebuilt: prebuilt nodes, can be nil
func LoadPlanFromJson(name string, jsonData []byte, prebuilt []Node) error {
	return Global.LoadPlanFromJson(name, jsonData, prebuilt)
}

type Engine struct {
	StateBuilder func() State

	builders map[string]BuildNodeFunc

	plans map[string]*Plan

	pools map[string]*_WorkerPool

	buildersLocker, plansLocker, poolsLocker sync.RWMutex
}

// RegisterNodeBuilder register node builder to engine
func (engine *Engine) RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	engine.buildersLocker.Lock()
	engine.builders[name] = builder
	engine.buildersLocker.Unlock()
}

// RegisterPlan register plan to engine
func (engine *Engine) RegisterPlan(name string, plan *Plan) error {
	err := plan.Init()
	if err != nil {
		return err
	}
	engine.plansLocker.Lock()
	engine.plans[name] = plan
	engine.plansLocker.Unlock()
	return nil
}

// LoadPlanFromJson load plan from json data
// name: name of plan to load
// jsonData: json data of plan
// prebuilt: prebuilt nodes, can be nil
func (engine *Engine) LoadPlanFromJson(name string, jsonData []byte, prebuilt []Node) error {
	plan := &Plan{}
	err := json.Unmarshal(jsonData, plan)
	if err != nil {
		return err
	}

	plan.prebuilt = make(map[string]Node)
	for _, node := range prebuilt {
		plan.prebuilt[node.Name()] = node
	}
	plan.version = strconv.FormatInt(time.Now().Unix(), 10)

	engine.plansLocker.Lock()
	engine.plans[name] = plan
	engine.plansLocker.Unlock()
	return nil
}

// ExecPlan exec plan register in engine
func (engine *Engine) ExecPlan(name string, ctx context.Context) <-chan Output {
	output := Output{}
	outputCh := make(chan Output, 1)

	if ctx == nil {
		ctx = context.Background()
	}

	go func() {
		engine.plansLocker.RLock()
		plan := engine.plans[name]
		engine.plansLocker.RUnlock()

		if plan == nil {
			output.Err = fmt.Errorf("plan not found, name: %s", name)
			outputCh <- output
			return
		}

		engine.poolsLocker.RLock()
		pool := engine.pools[name]
		engine.poolsLocker.RUnlock()

		// set worker pool for new plan
		if pool == nil {
			engine.poolsLocker.Lock()
			if engine.pools[name] == nil {
				engine.pools[name] = &_WorkerPool{
					Pool: sync.Pool{
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
			pool = engine.pools[name]
			engine.poolsLocker.Unlock()
		}

		// get worker from pool and work
		worker, err := pool.GetWorker()
		if err != nil {
			output.Err = err
			outputCh <- output
			return
		}
		output = <-worker.Work(ctx)
		outputCh <- output

		// if the plan has not been updated, reuse the worker
		plan.locker.RLock()
		version := plan.version
		plan.locker.RUnlock()
		if worker.Version == version {
			pool.PutWorker(worker)
		}
	}()

	return outputCh
}

// UpdatePlan update plan register in engine
func (engine *Engine) UpdatePlan(name string, update func(plan *Plan)) error {
	engine.plansLocker.RLock()
	plan := engine.plans[name]
	engine.plansLocker.RUnlock()

	plan.locker.Lock()
	update(plan)
	plan.locker.Unlock()

	err := plan.Init()
	if err != nil {
		return err
	}

	return nil
}

func (engine *Engine) ExportPlan(name string) ([]byte, error) {
	engine.plansLocker.RLock()
	plan := engine.plans[name]
	engine.plansLocker.RUnlock()

	if plan != nil {
		return json.Marshal(plan)
	} else {
		return nil, fmt.Errorf("plan: %s not found", name)
	}
}

// WarmupPool warm up pool to avoid cold start
// name: plan name
// size: set size of worker buf queue
func (engine *Engine) WarmupPool(name string, size int) {
	engine.poolsLocker.Lock()
	if engine.pools[name] != nil {
		engine.pools[name].Warmup(size)
	}
	engine.poolsLocker.Unlock()
}

// ClearPool clear worker pool of plan, invoke it to make plan effect immediately after update
// name: name of plan
func (engine *Engine) ClearPool(name string) {
	engine.poolsLocker.Lock()
	engine.pools[name] = nil
	engine.poolsLocker.Unlock()
}

func (engine *Engine) buildWorker(name string) (worker *_Worker, err error) {
	engine.plansLocker.RLock()
	plan := engine.plans[name]
	engine.plansLocker.RUnlock()

	plan.locker.RLock()
	defer plan.locker.RUnlock()

	nodeMap := map[string]Node{}
	reuse := map[string]Node{} // collect nodes which can be reused in the build nodes process

	for _, v := range plan.graph.Vertexes {
		nodeName := v.RefRoot.NodeName
		nodeMap[nodeName], err = engine.buildNode(plan, nodeName, "", reuse)
		if err != nil {
			return
		}
	}

	if len(reuse) > 0 {
		plan.locker.RUnlock()
		plan.locker.Lock()
		for nodeName, node := range reuse {
			plan.prebuilt[nodeName] = node
		}
		plan.locker.Unlock()
		plan.locker.RLock()
	}

	worker = &_Worker{
		Works:        newWorkList(plan.graph),
		Nodes:        nodeMap,
		StateBuilder: engine.StateBuilder,
		Version:      plan.version,
	}
	return
}

// buildNode build node by ref, props and prebuilt nodes.
// prefix will be added to node name,
// example: prefix = ClusterA, node name = SubNodeB => ClusterA.SubNodeB
func (engine *Engine) buildNode(plan *Plan, nodeName string, prefix string, reuse map[string]Node) (Node, error) {
	engine.buildersLocker.RLock()
	defer engine.buildersLocker.RUnlock()

	root := plan.graph.NodeRefs[nodeName]
	props := plan.props
	prebuilt := plan.prebuilt

	if props == nil {
		props = EmptyProps{}
	}

	var rootNode Node
	var err error

	if prefix != "" {
		nodeName = prefix + "." + root.NodeName
	} else {
		nodeName = root.NodeName
	}

	// prefer to use pre-built nodes
	if node := getPrebuiltNode(prebuilt, nodeName); node != nil {
		rootNode = node
	} else if builder := engine.builders[root.NodeType]; builder != nil {
		rootNode, err = builder(nodeName, props)
		if err != nil {
			return nil, fmt.Errorf("failed to build %s, err=%s", nodeName, err.Error())
		}
	} else {
		return nil, fmt.Errorf("no builder found for type %s", root.NodeType)
	}

	if root.ReUse {
		if _, ok := prebuilt[nodeName]; !ok {
			if cloneableNode, ok := rootNode.(Cloneable); ok {
				reuse[nodeName] = cloneableNode.Clone()
			} else {
				reuse[nodeName] = nil
			}
		}
	}

	// inject sub-nodes for cluster
	if cluster, ok := rootNode.(Cluster); ok {
		var subNodes []Node

		// build sub-nodes just like root node(cluster)
		for _, ref := range root.SubRefs {
			var subNode Node

			if len(ref.SubRefs) == 0 {
				if node := getPrebuiltNode(prebuilt, rootNode.Name()+"."+ref.NodeName); node != nil {
					subNode = node
				} else if builder := engine.builders[ref.NodeType]; builder != nil {
					subNode, err = builder(rootNode.Name()+"."+ref.NodeName, props)
					if err != nil {
						return nil, fmt.Errorf("failed to build %s, err=%s", nodeName, err.Error())
					}
				} else {
					return nil, fmt.Errorf("no builder found for type %s", ref.NodeType)
				}

				if ref.ReUse && prebuilt[rootNode.Name()+"."+ref.NodeName] == nil {
					reuse[rootNode.Name()+"."+ref.NodeName] = subNode
				}

				subNode, err = engine.wrapNode(subNode, ref.Wrappers, props)
				if err != nil {
					return nil, err
				}

				subNodes = append(subNodes, subNode)
			} else {
				if prefix == "" {
					prefix = root.NodeName
				} else {
					prefix = prefix + "." + root.NodeName
				}

				if subNode, err = engine.buildNode(plan, ref.NodeName, prefix, reuse); err != nil {
					return nil, err
				} else {
					subNodes = append(subNodes, subNode)
				}
			}
		}

		cluster.Inject(subNodes)
	}

	rootNode, err = engine.wrapNode(rootNode, root.Wrappers, props)
	if err != nil {
		return nil, err
	}

	return rootNode, nil
}

func (engine *Engine) wrapNode(target Node, wrappers []string, props Props) (Node, error) {
	for _, wrapper := range wrappers {
		if builder := engine.builders[wrapper]; builder != nil {
			node, err := builder(target.Name(), props)
			if err != nil {
				return nil, fmt.Errorf("failed to build %s, err=%s", wrapper, err.Error())
			}

			if wrapperNode, ok := node.(Wrapper); ok {
				wrapperNode.Wrap(target)
				target = wrapperNode
			}
		}
	}

	return target, nil
}

func getPrebuiltNode(prebuilt map[string]Node, nodeName string) Node {
	var node Node

	if prebuilt[nodeName] != nil {
		if cloneableNode, ok := prebuilt[nodeName].(Cloneable); ok {
			node = cloneableNode.Clone()
		}
	}

	return node
}
