package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func TestSimpleNode(t *testing.T) {
	x := 1

	running.RegisterNodeBuilder("Simple",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			x++
		}))

	running.RegisterNodeBuilder("SimpleState",
		common.NewSimpleStatefulNodeBuilder(func(ctx context.Context, state running.State) {
			state.Update("x", x)
		}))

	ops := []running.Option{
		running.AddNodes("Simple", "Simple1"),
		running.AddNodes("SimpleState", "Simple2"),
		running.LinkNodes("Simple1", "Simple2"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.RegisterPlan("TestSimpleNode", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestSimpleNode", context.Background())
	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		x := utils.ProxyState(output.State).GetInt("x")
		if x != 2 {
			t.Errorf("expect x = 2, but got %d", x)
		}
	}

	if x != 2 {
		t.Errorf("expect x = 2, but got %d", x)
	}
}
