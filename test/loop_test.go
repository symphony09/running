package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	_ "github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func TestLoopCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Loop", "L1", "L2"),
		running.AddNodes("BaseTest", "B1"),
		running.AddNodes("SetState", "S1"),
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

	err := running.Global.RegisterPlan("TestLoopCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("TestLoopCluster", context.Background())

	if out.Err != nil {
		t.Errorf("exec plan failed,err=%s", out.Err.Error())
	} else {
		sum := utils.GetRunSummary(out.State)
		t.Logf("plan cost %d ms", sum.Cost.Milliseconds())

		if len(sum.Logs["L1.B1"]) != 0 {
			t.Errorf("expect run 0 times, but got %d", len(sum.Logs["L1.B1"]))
			return
		}

		if len(sum.Logs["L2.B1"]) != 5 {
			t.Errorf("expect run 5 times, but got %d", len(sum.Logs["L2.B1"]))
			return
		}
	}
}
