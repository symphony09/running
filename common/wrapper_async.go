package common

import (
	"context"

	"github.com/symphony09/running"
)

type AsyncWrapper struct {
	running.BaseWrapper

	running.State

	PanicHandler func(ctx context.Context, nodeName string, v interface{})
}

func NewAsyncWrapper(name string, props running.Props) (running.Node, error) {
	wrapper := new(AsyncWrapper)

	handler, _ := props.SubGet(name, "panic_handler")
	if method, ok := handler.(func(ctx context.Context, nodeName string, v interface{})); ok {
		wrapper.PanicHandler = method
	}

	return wrapper, nil
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
		defer func() {
			if r := recover(); r != nil {
				if wrapper.PanicHandler != nil {
					wrapper.PanicHandler(ctx, node.Name(), r)
				}
			}
		}()

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
