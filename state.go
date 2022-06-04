package running

import (
	"sync"
)

func NewStandardState() *StandardState {
	return &StandardState{params: map[string]interface{}{}}
}

func NewChanState() *ChanState {
	state := &ChanState{ch: make(chan map[string]interface{}, 1)}
	state.ch <- map[string]interface{}{}
	return state
}

func NewUnsafeState() *UnsafeState {
	return &UnsafeState{params: map[string]interface{}{}}
}

type StandardState struct {
	sync.RWMutex

	params map[string]interface{}
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

type ChanState struct {
	ch chan map[string]interface{}
}

func (state *ChanState) Query(key string) (value interface{}, exists bool) {
	params := <-state.ch
	value, exists = params[key]
	state.ch <- params
	return
}

func (state *ChanState) Update(key string, value interface{}) {
	params := <-state.ch
	params[key] = value
	state.ch <- params
}

func (state *ChanState) Transform(key string, transform TransformStateFunc) {
	params := <-state.ch
	params[key] = transform(params[key])
	state.ch <- params
}

type UnsafeState struct {
	params map[string]interface{}
}

func (state *UnsafeState) Query(key string) (value interface{}, exists bool) {
	value, exists = state.params[key]
	return
}

func (state *UnsafeState) Update(key string, value interface{}) {
	state.params[key] = value
}

func (state *UnsafeState) Transform(key string, transform TransformStateFunc) {
	state.params[key] = transform(state.params[key])
}

type OverlayState struct {
	Upper State

	Lower State
}

// NewOverlayState return a OverlayState, isolate writes and updates
func NewOverlayState(lower, upper State) State {
	return OverlayState{
		Upper: upper,
		Lower: lower,
	}
}

func (state OverlayState) Query(key string) (value interface{}, exists bool) {
	if value, exists = state.Upper.Query(key); exists {
		return
	} else {
		value, exists = state.Lower.Query(key)
		if exists {
			state.Upper.Update(key, value)
		}
		return
	}
}

func (state OverlayState) Update(key string, value interface{}) {
	state.Upper.Update(key, value)
}

func (state OverlayState) Transform(key string, transform TransformStateFunc) {
	if value, exists := state.Upper.Query(key); !exists {
		if value, exists = state.Lower.Query(key); exists {
			state.Upper.Update(key, value)
		}
	}

	state.Upper.Transform(key, transform)
}
