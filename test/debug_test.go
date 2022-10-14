package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	_ "github.com/symphony09/running/common"
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
}
