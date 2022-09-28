package test

import (
	"context"
	"fmt"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

type BaseTestNode struct {
	running.Base
}

func (node *BaseTestNode) Run(ctx context.Context) {
	start := time.Now()

	select {
	case <-time.After(10 * time.Millisecond):
		utils.AddLog(node.State, node.Name(), start, time.Now(), "success", nil)
	case <-ctx.Done():
		utils.AddLog(node.State, node.Name(), start, time.Now(), "timeout", ctx.Err())
	}
}

type SetStateNode struct {
	running.Base

	key string

	value interface{}
}

func (node *SetStateNode) Run(ctx context.Context) {
	node.State.Update(node.key, node.value)
	utils.AddLog(node.State, node.Name(), time.Now(), time.Now(), "", nil)
}

type NothingNode struct {
	running.Base
}

func (node *NothingNode) Run(ctx context.Context) {}

type TimerWrapper struct {
	running.BaseWrapper
}

func (wrapper *TimerWrapper) Run(ctx context.Context) {
	start := time.Now()

	wrapper.Target.Run(ctx)

	fmt.Printf("Node %s cost %d ms\n", wrapper.Target.Name(), time.Since(start).Milliseconds())
}

type HighCostNode struct {
	running.Base
}

func (node *HighCostNode) Run(ctx context.Context) {
	fmt.Println(node.Name() + " start")
	time.Sleep(time.Second)
	fmt.Println(node.Name() + " end")
}

func (node *HighCostNode) Clone() running.Node {
	return &HighCostNode{
		Base: running.Base{
			NodeName: node.Name() + "_clone",
		},
	}
}
