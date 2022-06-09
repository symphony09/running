package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func TestSerialization(t *testing.T) {
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

	data, err := json.Marshal(plan)
	if err != nil {
		t.Errorf("marshal plan failed, err=%s", err.Error())
	}

	err = running.LoadPlanFromJson("TestSerialization", data, nil)
	if err != nil {
		t.Errorf("load plan failed, err=%s", err.Error())
	}

	output := <-running.ExecPlan("TestSerialization", context.Background())
	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		sum := utils.GetRunSummary(output.State)
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
