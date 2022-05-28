package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func TestSwitchCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Switch", "Switch1", "Switch2", "Switch3"),
		running.AddNodes("BaseTest", "Base1", "Base2"),
		running.AddNodes("SetState", "State1"),
		running.MergeNodes("Switch1", "Base1", "Base2"),
		running.MergeNodes("Switch2", "Base1", "Base2"),
		running.MergeNodes("Switch3", "Base1", "Base2"),
		running.LinkNodes("State1", "Switch1", "Switch2", "Switch3"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Switch2.status": "on",
		"Switch3.watch":  "status?",
		"State1.key":     "status?",
		"State1.value":   "on",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("TestSwitchCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("TestSwitchCluster", context.Background())

	if out.Err != nil {
		t.Errorf("exec plan failed,err=%s", out.Err.Error())
	} else {
		sum := utils.GetRunSummary(out.State)

		if len(sum.Logs["Switch1.Base1"]) != 0 {
			t.Errorf("expect Switch1.Base1 run 0 times, but got %d", len(sum.Logs["Switch1.Base1"]))
		}

		if len(sum.Logs["Switch1.Base2"]) != 0 {
			t.Errorf("expect Switch1.Base2 run 0 times, but got %d", len(sum.Logs["Switch1.Base2"]))
		}

		if len(sum.Logs["Switch2.Base1"]) == 0 {
			t.Error("expect Switch2.Base1 run 1 times, but got 0")
		}

		if len(sum.Logs["Switch2.Base2"]) == 0 {
			t.Error("expect Switch2.Base2 run 1 times, but got 0")
		}

		if len(sum.Logs["Switch3.Base1"]) == 0 {
			t.Error("expect Switch3.Base1 run 1 times, but got 0")
		}

		if len(sum.Logs["Switch3.Base2"]) == 0 {
			t.Error("expect Switch3.Base2 run 1 times, but got 0")
		}
	}
}
