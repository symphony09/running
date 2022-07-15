package common

import "github.com/symphony09/running"

type OverlayState struct {
	Upper running.State

	Lower running.State
}

// NewOverlayState return a OverlayState, isolate writes and updates
func NewOverlayState(lower, upper running.State) running.State {
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

func (state OverlayState) Transform(key string, transform running.TransformStateFunc) {
	if value, exists := state.Upper.Query(key); !exists {
		if value, exists = state.Lower.Query(key); exists {
			state.Upper.Update(key, value)
		}
	}

	state.Upper.Transform(key, transform)
}
