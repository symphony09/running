package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
)

func TestSimpleNode(t *testing.T) {
	x := 1

	running.Global.RegisterNodeBuilder("Simple",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			x++
		}))

	ops := []running.Option{
		running.AddNodes("Simple", "Simple1"),
		running.LinkNodes("Simple1"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.Global.RegisterPlan("TestSimpleNode", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	<-running.Global.ExecPlan("TestSimpleNode", context.Background())

	if x != 2 {
		t.Errorf("expect x = 2, but got %d", x)
	}
}
