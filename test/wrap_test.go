package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
)

func TestWrap(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Loop", "L1", "L2"),
		running.AddNodes("BaseTest", "B1"),
		running.AddNodes("SetState", "S1"),
		running.WrapAllNodes("TimerWrapper"),
		running.MergeNodes("L1", "B1"),
		running.MergeNodes("L2", "B1"),
		running.SLinkNodes("L1", "S1", "L2"),
	}

	props := running.StandardProps(map[string]interface{}{
		"L1.max_loop": 5,
		"L2.max_loop": 5,
		"L1.watch":    "loop?",
		"L2.watch":    "loop?",
		"S1.key":      "loop?",
		"S1.value":    true,
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestWrap", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestWrap", context.Background())
	if output.Err != nil {
		t.Errorf("exec plan failed, err=%s", output.Err.Error())
		return
	}
}
