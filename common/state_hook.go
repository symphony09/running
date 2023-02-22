package common

import "github.com/symphony09/running"

type HookState struct {
	Base running.State

	Hooks []func(state running.State, event, key string, completed bool)
}

func NewHookState(base running.State, hooks ...func(state running.State, event, key string, completed bool)) *HookState {
	if base == nil {
		base = running.NewStandardState()
	}

	return &HookState{Base: base, Hooks: hooks}
}

func (state *HookState) Query(key string) (interface{}, bool) {
	for _, hook := range state.Hooks {
		hook(state.Base, "query", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "query", key, true)
		}
	}()

	return state.Base.Query(key)
}

func (state *HookState) Update(key string, value interface{}) {
	for _, hook := range state.Hooks {
		hook(state.Base, "update", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "update", key, true)
		}
	}()

	state.Base.Update(key, value)
}

func (state *HookState) Transform(key string, transform running.TransformStateFunc) {
	for _, hook := range state.Hooks {
		hook(state.Base, "transform", key, false)
	}

	defer func() {
		for _, hook := range state.Hooks {
			hook(state.Base, "transform", key, true)
		}
	}()

	state.Base.Transform(key, transform)
}
