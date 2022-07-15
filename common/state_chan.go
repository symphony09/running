package common

import "github.com/symphony09/running"

type ChanState struct {
	ch chan map[string]interface{}
}

func NewChanState() *ChanState {
	state := &ChanState{ch: make(chan map[string]interface{}, 1)}
	state.ch <- map[string]interface{}{}
	return state
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

func (state *ChanState) Transform(key string, transform running.TransformStateFunc) {
	params := <-state.ch
	params[key] = transform(params[key])
	state.ch <- params
}
