package test

import (
	"context"
	"testing"

	"running"
	"running/utils"
)

func TestSelectCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Select", "Select1", "Select2"),
		running.AddNodes("BaseTest", "Base1", "Base2"),
		running.AddNodes("SetState", "State1"),
		running.MergeNodes("Select1", "Base1", "Base2"),
		running.MergeNodes("Select2", "Base1", "Base2"),
		running.SLinkNodes("Select1", "State1", "Select2"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Select1.selected": "Base1",
		"Select1.watch":    "select?",
		"Select2.selected": "Base1",
		"Select2.watch":    "select?",
		"State1.key":       "select?",
		"State1.value":     "Base2",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("TestSelectCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("TestSelectCluster", context.Background())

	if out.Err != nil {
		t.Errorf("exec plan failed,err=%s", out.Err.Error())
	} else {
		sum := utils.GetRunSummary(out.State)

		if len(sum.Logs["Select1.Base1"]) == 0 {
			t.Error("expect Select1.Base1 run 1 times, but got 0")
		}

		if len(sum.Logs["Select1.Base2"]) != 0 {
			t.Errorf("expect Select1.Base2 run 0 times, but got %d", len(sum.Logs["Select1.Base2"]))
		}

		if len(sum.Logs["Select2.Base1"]) != 0 {
			t.Errorf("expect Select2.Base1 run 0 times, but got %d", len(sum.Logs["Select2.Base1"]))
		}

		if len(sum.Logs["Select2.Base2"]) == 0 {
			t.Error("expect Select2.Base2 run 1 times, but got 0")
		}
	}
}
