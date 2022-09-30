package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func TestTransactionalCluster(t *testing.T) {
	running.RegisterNodeBuilder("Boom",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			panic("boom")
		}))

	ops := []running.Option{
		running.AddNodes("SetState", "S1", "S2"),
		running.AddNodes("Boom", "B1"),
		running.AddNodes("Transactional", "T1"),
		running.MergeNodes("T1", "S2", "B1"),
		running.SLinkNodes("S1", "T1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"S1.key":      "stage",
		"S1.value":    "1",
		"T1.S2.key":   "stage",
		"T1.S2.value": "2",
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestTransactionalCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.ExecPlan("TestTransactionalCluster", context.Background())

	if out.Err == nil {
		t.Error("exec plan succeeded unexpectedly")
	} else {
		sum := utils.GetRunSummary(out.State)

		if len(sum.Logs["T1.S2"]) != 2 || sum.Logs["T1.S2"][1].Msg != "value reverted" {
			t.Error("revert message not found")
		}

		helper := utils.ProxyState(out.State)

		if helper.GetString("stage") != "1" {
			t.Errorf("value not reverted, got:%s", helper.GetString("stage"))
		}
	}
}
