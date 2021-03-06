package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func TestBase(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("BaseTest", "B1", "B2", "B3", "B4", "B5", "B6"),
		running.AddNodes("SetState", "S1"),
		running.LinkNodes("B1", "B2", "B3"),
		running.SLinkNodes("B3", "B4", "B5"),
		running.RLinkNodes("S1", "B2", "B5"),
	}

	props := running.StandardProps(map[string]interface{}{
		"S1.key":   "test_key",
		"S1.value": "test_value",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("Base", plan)
	if err != nil {
		t.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	output := <-running.ExecPlan("Base", ctx)
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
	if sum.Logs["B1"][0].End.After(sum.Logs["B2"][0].Start) {
		t.Error("B2 start before B1 end")
	}
	if sum.Logs["B2"][0].End.Before(sum.Logs["B3"][0].Start) {
		t.Error("B2 end before B3 start")
	}
	if sum.Logs["B1"][0].Msg != "success" {
		t.Errorf("expect B1 success, but got %s", sum.Logs["B5"][0].Msg)
	}
	if sum.Logs["B5"][0].Msg != "timeout" {
		t.Errorf("expect B5 timeout, but got %s", sum.Logs["B5"][0].Msg)
	}
	if len(sum.Logs["B6"]) != 0 {
		t.Errorf("expect B6 run count = 0, but got %d", len(sum.Logs["B6"]))
	}

	err = running.UpdatePlan("Base", func(plan *running.Plan) {
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

	output = <-running.ExecPlan("Base", ctx)
	if value := utils.ProxyState(output.State).GetString("test_key"); value != "test_value" {
		t.Errorf("expect state value = test_value, but got %s", value)
	}

	err = running.UpdatePlan("Base", func(plan *running.Plan) {
		plan.Props = running.StandardProps(map[string]interface{}{
			"S1.key":   "test_key",
			"S1.value": "test_value2",
		})

		plan.Options[len(plan.Options)-1] = running.SLinkNodes("S1", "B6")
		return
	})

	running.ClearPool("Base")

	if err != nil {
		t.Errorf("update plan failed, expect success, err=%s", err.Error())
		return
	}

	output = <-running.ExecPlan("Base", ctx)
	if value := utils.ProxyState(output.State).GetString("test_key"); value != "test_value2" {
		t.Errorf("expect state value = test_value2, but got %s", value)
	}

	sum = utils.GetRunSummary(output.State)
	if len(sum.Logs["B6"]) == 0 {
		t.Error("expect B6 run count != 0, but got 0")
	}
}

func TestOverlayState(t *testing.T) {
	upper, lower := running.NewStandardState(), running.NewStandardState()
	overlay := common.NewOverlayState(lower, upper)
	helper1, helper2 := utils.ProxyState(overlay), utils.ProxyState(lower)

	lower.Update("a", 1)
	overlay.Update("b", 1)
	lower.Update("c", 1)
	overlay.Transform("c", func(from interface{}) interface{} {
		x, _ := from.(int)
		return x + 1
	})

	if helper1.GetInt("a") != 1 {
		t.Errorf("expect a = 1, but got %d", helper1.GetInt("a"))
	}

	if helper2.GetInt("a") != 1 {
		t.Errorf("expect a = 1, but got %d", helper1.GetInt("a"))
	}

	if helper1.GetInt("b") != 1 {
		t.Errorf("expect b = 1, but got %d", helper1.GetInt("b"))
	}

	if helper2.GetInt("b") != 0 {
		t.Errorf("expect b = 0, but got %d", helper1.GetInt("b"))
	}

	if helper1.GetInt("c") != 2 {
		t.Errorf("expect c = 2, but got %d", helper1.GetInt("c"))
	}

	if helper2.GetInt("c") != 1 {
		t.Errorf("expect c = 1, but got %d", helper1.GetInt("c"))
	}
}

func TestPanic(t *testing.T) {
	running.RegisterNodeBuilder("Base", func(name string, props running.Props) (running.Node, error) {
		return &running.Base{}, nil
	})

	ops := []running.Option{
		running.AddNodes("Base", "B1"),
		running.AddNodes("BaseTest", "T1"),
		running.SLinkNodes("B1", "T1"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.RegisterPlan("TestPanic", plan)
	if err != nil {
		t.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	output := <-running.ExecPlan("TestPanic", context.Background())
	if output.Err != nil {
		fmt.Printf("exec plan failed, err=%s\n", output.Err.Error())

		sum := utils.GetRunSummary(output.State)
		if len(sum.Logs["T1"]) != 0 {
			t.Errorf("expect T1 run count = 0, but got %d", len(sum.Logs["T1"]))
		}
	} else {
		t.Errorf("exec plan successfully")
	}
}

func TestReUseNodes(t *testing.T) {
	count := 0
	running.RegisterNodeBuilder("BuildCounter", func(name string, props running.Props) (running.Node, error) {
		count++
		node := new(NothingNode)
		node.SetName(name)
		return node, nil
	})

	plan := running.NewPlan(nil, nil,
		running.AddNodes("BuildCounter", "B1"),
		running.ReUseNodes("B1"),
		running.LinkNodes("B1"))

	err := running.RegisterPlan("TestReUseNodes", plan)
	if err != nil {
		t.Error(err)
		return
	}

	<-running.ExecPlan("TestReUseNodes", nil)
	running.ClearPool("TestReUseNodes")
	<-running.ExecPlan("TestReUseNodes", nil)

	if count != 1 {
		t.Errorf("expect build count = 1, but got %d", count)
	}
}

func BenchmarkExecPlan(b *testing.B) {
	ops := []running.Option{
		running.AddNodes("Nothing", "N1", "N2", "N3", "N4"),
		running.LinkNodes("N1", "N4"),
		running.SLinkNodes("N1", "N2", "N3"),
		running.ReUseNodes("N1", "N2", "N3", "N4"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.RegisterPlan("BenchmarkExecPlan", plan)
	if err != nil {
		b.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	for i := 0; i < b.N; i++ {
		running.ExecPlan("BenchmarkExecPlan", nil)
	}
}
