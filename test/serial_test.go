package test

import (
	"context"
	"testing"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func TestSerialCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Serial", "Serial1"),
		running.AddNodes("BaseTest", "Base1", "Base2", "Base3"),
		running.MergeNodes("Serial1", "Base1", "Base2", "Base3"),
		running.SLinkNodes("Serial1", "END"),
	}

	plan := running.NewPlan(running.EmptyProps{}, nil, ops...)

	err := running.RegisterPlan("TestSerialCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestSerialCluster", context.Background())

	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		sum := utils.GetRunSummary(output.State)
		if sum.Logs["Serial1.Base1"][0].End.After(sum.Logs["Serial1.Base2"][0].Start) {
			t.Error("Base2 start before Base1 end")
		}
		if sum.Logs["Serial1.Base2"][0].End.After(sum.Logs["Serial1.Base3"][0].Start) {
			t.Error("Base3 start before Base2 end")
		}
	}
}
