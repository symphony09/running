package test

import (
	"context"
	"errors"
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
		t.Errorf("expect B1 success, but got %s", sum.Logs["B1"][0].Msg)
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

		if !errors.Is(output.Err, running.ErrWorkerPanic) {
			t.Errorf("expect WorkPanicError, but got %t", err)
		}

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

func TestSkipNodes(t *testing.T) {
	plan := running.NewPlan(nil, nil,
		running.AddNodes("HighCost", "H1"),
		running.AddNodes("BaseTest", "B1", "B2"),
		running.WrapAllNodes("Debug"),
		running.RLinkNodes("B2", "B1", "H1"))

	err := running.RegisterPlan("TestSkipNodes", plan)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.WithValue(context.Background(), running.CtxKey, running.CtxParams{SkipNodes: []string{"H1"}})

	output := <-running.ExecPlan("TestSkipNodes", ctx)
	if output.Err != nil {
		t.Error(output.Err)
	}

	sum := utils.GetRunSummary(output.State)
	if len(sum.Logs["B1"]) != 1 {
		t.Errorf("expect B1 run count eq 1, but got %d", len(sum.Logs["B1"]))
	}
	if len(sum.Logs["B2"]) != 1 {
		t.Errorf("expect B2 run count eq 1, but got %d", len(sum.Logs["B2"]))
	}
	if len(sum.Logs["H1"]) != 0 {
		t.Errorf("expect N1 run count eq 0, but got %d", len(sum.Logs["H1"]))
	}
}

func TestMarkNodes(t *testing.T) {
	plan := running.NewPlan(nil, nil,
		running.AddNodes("BaseTest", "B1", "B2", "B3", "B4", "B5", "B6"),
		running.MarkNodes("group1", "B1", "B2", "B3", "B4"),
		running.MarkNodes("group2", "B1", "B2"),
		running.MarkNodes("group3", "B1", "B3", "B5"),
		running.MarkNodes("group4", "B4", "B5"),
		running.WrapAllNodes("Debug"),
		running.LinkNodes("B1", "B2", "B3", "B4", "B5", "B6"))

	err := running.RegisterPlan("TestMarkNodes", plan)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.WithValue(context.Background(), running.CtxKey,
		running.CtxParams{
			MatchAllLabels:   []string{"group1"},
			MatchOneOfLabels: []string{"group2", "group3"},
		})

	output := <-running.ExecPlan("TestMarkNodes", ctx)
	if output.Err != nil {
		t.Error(output.Err)
	}

	// ( group1 ∩ group2 ) ∪ ( group1 ∩ group3 ) ∪ no labels = B1, B2, B3, B6
	sum := utils.GetRunSummary(output.State)
	if len(sum.Logs["B1"]) != 1 {
		t.Errorf("expect B1 run count eq 1, but got %d", len(sum.Logs["B1"]))
	}
	if len(sum.Logs["B2"]) != 1 {
		t.Errorf("expect B2 run count eq 1, but got %d", len(sum.Logs["B2"]))
	}
	if len(sum.Logs["B3"]) != 1 {
		t.Errorf("expect B3 run count eq 1, but got %d", len(sum.Logs["B3"]))
	}
	if len(sum.Logs["B6"]) != 1 {
		t.Errorf("expect B6 run count eq 1, but got %d", len(sum.Logs["B6"]))
	}
	if len(sum.Logs["B4"]) != 0 {
		t.Errorf("expect B4 run count eq 0, but got %d", len(sum.Logs["B4"]))
	}
	if len(sum.Logs["B5"]) != 0 {
		t.Errorf("expect B5 run count eq 0, but got %d", len(sum.Logs["B5"]))
	}
}

func TestCtxWithState(t *testing.T) {
	e := running.NewDefaultEngine()
	e.RegisterNodeBuilder("Simple", common.NewSimpleStatefulNodeBuilder(func(ctx context.Context, state running.State) {
		state.Update("data", 1)
		state.Transform("data", func(from interface{}) interface{} {
			return 2
		})
		state.Query("data")
	}))

	version := 0

	state := common.NewHookState(nil, func(state running.State, event, key string, completed bool) {
		if completed && event != "query" {
			version++
		}
	})

	ctx := context.WithValue(context.Background(), running.CtxKey, running.CtxParams{State: state})

	plan := running.NewPlan(nil, nil,
		running.AddNodes("Simple", "S"),
		running.LinkNodes("S", "END"))

	err := e.RegisterPlan("TestCtxWithState", plan)
	if err != nil {
		t.Error(err)
	}

	<-e.ExecPlan("TestCtxWithState", ctx)

	if version != 2 {
		t.Error(fmt.Errorf("expect version eq 2, but got %d", version))
	}
}

func TestSkipOnCtxErr(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("BaseTest", "B1", "B2"),
		running.AddNodes("SetState", "S1"),
		running.SLinkNodes("B1", "B2", "S1"),
		running.WrapAllNodes("Debug"),
	}

	props := running.StandardProps(map[string]interface{}{
		"S1.key":   "k",
		"S1.value": "v",
	})

	plan := running.NewPlan(props, nil, ops...)

	running.RegisterPlan("TestSkipOnCtxErr", plan)

	running.WarmupPool("TestSkipOnCtxErr", 1)

	ctx := context.WithValue(context.Background(), running.CtxKey,
		running.CtxParams{
			SkipOnCtxErr: true,
		})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	output := <-running.ExecPlan("TestSkipOnCtxErr", ctx)
	if !errors.Is(output.Err, context.DeadlineExceeded) {
		t.Errorf("expect deadline exceeded error, bug got %v", output.Err)
	}

	sum := utils.GetRunSummary(output.State)
	if len(sum.Logs["B1"]) != 1 {
		t.Errorf("expect B1 run count eq 1, but got %d", len(sum.Logs["B1"]))
	}
	if len(sum.Logs["B2"]) != 0 {
		t.Errorf("expect B2 run count eq 0, but got %d", len(sum.Logs["B2"]))
	}

	if v, _ := output.State.Query("k"); v != nil {
		t.Error("expect got empty v")
	}
}

func TestVirtualNodes(t *testing.T) {
	ops := []running.Option{
		running.AddVirtualNodes("begin", "stage1", "end"),
		running.AddNodes("Nothing", "N1", "N2", "N3", "N4"),
		running.LinkNodes("begin", "N1", "N2"),
		running.RLinkNodes("stage1", "N1", "N2"),
		running.LinkNodes("stage1", "N3", "N4"),
		running.RLinkNodes("end", "N3", "N4"),

		// begin -> N1, N2 -> stage1 -> N1, N4 -> end

		running.WrapAllNodes("Debug"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.RegisterPlan("TestVirtualNodes", plan)
	if err != nil {
		t.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	ouput := <-running.ExecPlan("TestVirtualNodes", nil)
	if ouput.Err != nil {
		t.Error(ouput.Err)
	}
}

func init() {
	ops := []running.Option{
		running.AddNodes("Nothing", "N1", "N2", "N3", "N4"),
		running.LinkNodes("N1", "N4"),
		running.SLinkNodes("N1", "N2", "N3"),
		running.ReUseNodes("N1", "N2", "N3", "N4"),
	}

	plan := running.NewPlan(nil, nil, ops...)

	err := running.RegisterPlan("BenchmarkExecPlan", plan)
	if err != nil {
		panic(fmt.Errorf("register plan failed, err=%s", err.Error()))
		return
	}

	running.WarmupPool("BenchmarkExecPlan", 100)
}

func BenchmarkExecPlan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		running.ExecPlan("BenchmarkExecPlan", nil)
	}
}
