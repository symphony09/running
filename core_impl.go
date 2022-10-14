package running

import (
	"context"
	"encoding/json"
	"sync"
)

// Base a simple impl of Node, Cluster, Stateful
// Embed it in custom node and override interface methods as needed
type Base struct {
	NodeName string

	State State

	SubNodes []Node

	SubNodesMap map[string]Node
}

func (base *Base) SetName(name string) {
	base.NodeName = name
}

func (base *Base) Name() string {
	return base.NodeName
}

func (base *Base) Inject(nodes []Node) {
	base.SubNodes = append(base.SubNodes, nodes...)

	if base.SubNodesMap == nil {
		base.SubNodesMap = make(map[string]Node)
	}

	for _, node := range nodes {
		base.SubNodesMap[node.Name()] = node
	}
}

func (base *Base) Bind(state State) {
	base.State = state

	for _, node := range base.SubNodes {
		if statefulNode, ok := node.(Stateful); ok {
			statefulNode.Bind(state)
		}
	}
}

func (base *Base) Run(ctx context.Context) {
	panic("please implement run method")
}

func (base *Base) Reset() {
	base.State = nil
	base.ResetSubNodes()
}

func (base *Base) ResetSubNodes() {
	for _, node := range base.SubNodes {
		node.Reset()
	}
}

func (base *Base) Revert(ctx context.Context) {
	for _, node := range base.SubNodes {
		if reversibleNode, ok := node.(Reversible); ok {
			reversibleNode.Revert(ctx)
		}
	}
}

type BaseWrapper struct {
	Target Node

	State State
}

func (wrapper *BaseWrapper) Wrap(target Node) {
	wrapper.Target = target
}

func (wrapper *BaseWrapper) Name() string {
	return wrapper.Target.Name()
}

func (wrapper *BaseWrapper) Run(ctx context.Context) {
	wrapper.Target.Run(ctx)
}

func (wrapper *BaseWrapper) Reset() {
	wrapper.State = nil

	wrapper.Target.Reset()
}

func (wrapper *BaseWrapper) Bind(state State) {
	wrapper.State = state

	if statefulTarget, ok := wrapper.Target.(Stateful); ok {
		statefulTarget.Bind(state)
	}
}

type StandardProps map[string]interface{}

func (props StandardProps) Get(key string) (value interface{}, exists bool) {
	value, exists = props[key]
	return
}

func (props StandardProps) SubGet(sub, key string) (value interface{}, exists bool) {
	return props.Get(sub + "." + key)
}

func (props StandardProps) Copy() Props {
	cp := make(map[string]interface{})
	for k, v := range props {
		cp[k] = v
	}

	return StandardProps(cp)
}

func (props StandardProps) MarshalJSON() ([]byte, error) {
	propsMap := map[string]interface{}(props)
	return json.Marshal(propsMap)
}

type EmptyProps struct{}

func (props EmptyProps) Get(key string) (value interface{}, exists bool) {
	return
}

func (props EmptyProps) SubGet(sub, key string) (value interface{}, exists bool) {
	return
}

func (props EmptyProps) Copy() Props {
	return EmptyProps{}
}

type StandardState struct {
	sync.RWMutex

	params map[string]interface{}
}

func NewStandardState() *StandardState {
	return &StandardState{params: map[string]interface{}{}}
}

func (state *StandardState) Query(key string) (value interface{}, exists bool) {
	state.RLock()
	value, exists = state.params[key]
	state.RUnlock()
	return
}

func (state *StandardState) Update(key string, value interface{}) {
	state.Lock()
	state.params[key] = value
	state.Unlock()
}

func (state *StandardState) Transform(key string, transform TransformStateFunc) {
	state.Lock()
	state.params[key] = transform(state.params[key])
	state.Unlock()
}
