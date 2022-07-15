package common

import "github.com/symphony09/running"

type UnsafeState struct {
	params map[string]interface{}
}

func NewUnsafeState() *UnsafeState {
	return &UnsafeState{params: map[string]interface{}{}}
}

func (state *UnsafeState) Query(key string) (value interface{}, exists bool) {
	value, exists = state.params[key]
	return
}

func (state *UnsafeState) Update(key string, value interface{}) {
	state.params[key] = value
}

func (state *UnsafeState) Transform(key string, transform running.TransformStateFunc) {
	state.params[key] = transform(state.params[key])
}
