package common

import (
	"context"

	"github.com/symphony09/running"
)

type AsyncWrapper struct {
	running.BaseWrapper

	running.State
}

func NewAsyncWrapper(name string, props running.Props) (running.Node, error) {
	return new(AsyncWrapper), nil
}

func (wrapper *AsyncWrapper) Run(ctx context.Context) {
	var node running.Node
	if cloneableTarget, ok := wrapper.Target.(running.Cloneable); ok {
		node = cloneableTarget.Clone()
	} else {
		node = wrapper.Target
	}

	if statefulNode, ok := node.(running.Stateful); ok {
		statefulNode.Bind(wrapper.State)
	}

	go func() {
		node.Run(ctx)
		node.Reset()
	}()
}

func (wrapper *AsyncWrapper) Bind(state running.State) {
	wrapper.State = NewOverlayState(state, running.NewStandardState())
}

func (wrapper *AsyncWrapper) Reset() {
	return
}
