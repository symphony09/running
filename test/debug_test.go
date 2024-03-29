package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
)

func TestDebug(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("SetState", "S1", "S2"),
		running.WrapNodes("Debug", "S1", "S2"),
		running.SLinkNodes("S1", "S2"),
	}

	props := running.StandardProps(map[string]interface{}{
		"S1.key":   "val_int",
		"S1.value": 1,
		"S1.debug": "val_ctx,val_int,val_str",

		"S2.key":   "val_str",
		"S2.value": "test",
		"S2.debug": "ctx:val_ctx, state_in:val_int, state_out:val_str",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestDebug", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestDebug", context.WithValue(context.Background(), "val_ctx", true))
	if output.Err != nil {
		t.Errorf("exec plan failed, err=%s", output.Err.Error())
		return
	}

	running.RegisterNodeBuilder("OutOfRange", common.NewSimpleNodeBuilder(func(ctx context.Context) {
		a := make([]int, 0)
		a[1] = 1
	}))

	err = running.UpdatePlan("TestDebug", func(plan *running.Plan) {
		plan.Options = []running.Option{
			running.AddNodes("OutOfRange", "O1"),
			running.WrapNodes("Debug", "O1"),
			running.SLinkNodes("S2", "O1"),
		}
	})
	if err != nil {
		t.Errorf("update plan failed, err=%s", err.Error())
		return
	}

	running.ClearPool("TestDebug")

	output = <-running.ExecPlan("TestDebug", context.WithValue(context.Background(), "val_ctx", true))
	if output.Err == nil {
		t.Error("exec plan succeeded unexpectedly")
		return
	}
}
