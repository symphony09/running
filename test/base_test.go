package test

import (
	"context"
	"testing"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func TestBase(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("BaseTest", "B1", "B2", "B3", "B4", "B5", "B6"),
		running.AddNodes("SetState", "S1"),
		running.LinkNodes("B1", "B2", "B3"),
		running.SLinkNodes("B3", "B4", "B5", "S1"),
		running.SLinkNodes("B2", "S1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"S1.key":   "test_key",
		"S1.value": "test_value",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("Base", plan)
	if err != nil {
		t.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	output := <-running.Global.ExecPlan("Base", ctx)
	if output.Err != nil {
		t.Errorf("exec plan failed, err=%s", output.Err.Error())
		return
	}

	if value := utils.ProxyState(output.State).GetString("test_key"); value != "test_value" {
		t.Errorf("expect state value = test_value, but got %s", value)
	}

	sum := utils.GetRunSummary(output.State)
	if sum.Count != 6 {
		t.Errorf("expect run count = 6, but got %d", sum.Count)
	}
	if !sum.Logs["B1"][0].End.Before(sum.Logs["B2"][0].Start) {
		t.Error("B2 start before B1 end")
	}
	if !sum.Logs["B2"][0].End.After(sum.Logs["B3"][0].Start) {
		t.Error("B2 end before B3 start")
	}
	if sum.Logs["B4"][0].Msg != "success" {
		t.Errorf("expect B4 success, but got %s", sum.Logs["B5"][0].Msg)
	}
	if sum.Logs["B5"][0].Msg != "timeout" {
		t.Errorf("expect B5 timeout, but got %s", sum.Logs["B5"][0].Msg)
	}
	if len(sum.Logs["B6"]) != 0 {
		t.Errorf("expect B6 run count = 0, but got %d", len(sum.Logs["B6"]))
	}

	err = running.Global.UpdatePlan("Base", true, func(plan *running.Plan) {
		plan.Props = running.StandardProps(map[string]interface{}{
			"S1.key":   "test_key",
			"S1.value": "test_value2",
		})

		plan.Options = append(plan.Options, running.SLinkNodes("S1", "B1"))
		return
	})

	if err == nil {
		t.Error("update plan success, expect failed")
		return
	}

	output = <-running.Global.ExecPlan("Base", ctx)
	if value := utils.ProxyState(output.State).GetString("test_key"); value != "test_value" {
		t.Errorf("expect state value = test_value, but got %s", value)
	}

	err = running.Global.UpdatePlan("Base", true, func(plan *running.Plan) {
		plan.Props = running.StandardProps(map[string]interface{}{
			"S1.key":   "test_key",
			"S1.value": "test_value2",
		})

		plan.Options[len(plan.Options)-1] = running.SLinkNodes("S1", "B6")
		return
	})

	if err != nil {
		t.Errorf("update plan failed, expect success, err=%s", err.Error())
		return
	}

	output = <-running.Global.ExecPlan("Base", ctx)
	if value := utils.ProxyState(output.State).GetString("test_key"); value != "test_value2" {
		t.Errorf("expect state value = test_value2, but got %s", value)
	}

	sum = utils.GetRunSummary(output.State)
	if len(sum.Logs["B6"]) == 0 {
		t.Error("expect B6 run count != 0, but got 0")
	}
}
