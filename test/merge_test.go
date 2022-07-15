package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func TestMergeCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Merge", "Merge1"),
		running.AddNodes("SetState", "State1", "State2"),
		running.MergeNodes("Merge1", "State1", "State2"),
		running.SLinkNodes("Merge1", "END"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Merge1.merge": func(state, subState running.State) {
			helper := utils.ProxyState(subState)
			result := helper.GetInt("result")

			state.Transform("result", func(from interface{}) interface{} {
				sum, _ := from.(int)
				return sum + result
			})
		},
		"Merge1.State1.key":   "result",
		"Merge1.State1.value": 1,
		"Merge1.State2.key":   "result",
		"Merge1.State2.value": 2,
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestMergeCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestMergeCluster", context.Background())

	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		helper := utils.ProxyState(output.State)
		result := helper.GetInt("result")
		if result != 3 {
			t.Errorf("expect merge result = 3, but got %d", result)
		}
	}
}
