package common

import (
	"context"

	"github.com/symphony09/running"
)

type SimpleNode struct {
	running.Base

	Handler func(ctx context.Context)

	WithStateHandler func(ctx context.Context, state running.State)
}

func NewSimpleNodeBuilder(handler func(ctx context.Context)) running.BuildNodeFunc {
	return func(name string, props running.Props) (running.Node, error) {
		node := &SimpleNode{Handler: handler}
		node.SetName(name)

		return node, nil
	}
}

func NewSimpleStatefulNodeBuilder(handler func(ctx context.Context, state running.State)) running.BuildNodeFunc {
	return func(name string, props running.Props) (running.Node, error) {
		node := &SimpleNode{WithStateHandler: handler}
		node.SetName(name)

		return node, nil
	}
}

func (node *SimpleNode) Run(ctx context.Context) {
	if node.WithStateHandler != nil {
		node.WithStateHandler(ctx, node.State)
		return
	}

	if node.Handler != nil {
		node.Handler(ctx)
	}
}
