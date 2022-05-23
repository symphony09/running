package test

import (
	"context"
	"testing"
	"time"

	"running"
	"running/common"
	"running/utils"
)

func TestAspectCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Aspect", "Aspect1"),
		running.AddNodes("BaseTest", "Base1"),
		running.MergeNodes("Aspect1", "Base1"),
		running.SLinkNodes("Aspect1", "END"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Aspect1.around": func(cluster *common.AspectCluster) {
			current := time.Now()
			logName := cluster.Name() + ".Around"
			utils.AddLog(cluster.State, logName, current, current, "", nil)
		},
		"Aspect1.before": func(cluster *common.AspectCluster) {
			current := time.Now()
			logName := cluster.Name() + ".Before"
			utils.AddLog(cluster.State, logName, current, current, "", nil)
		},
		"Aspect1.after": func(cluster *common.AspectCluster) {
			current := time.Now()
			logName := cluster.Name() + ".After"
			utils.AddLog(cluster.State, logName, current, current, "", nil)
		},
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("TestAspectCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.Global.ExecPlan("TestAspectCluster", context.Background())

	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		sum := utils.GetRunSummary(output.State)
		if !sum.Logs["Aspect1.Around"][0].End.Before(sum.Logs["Aspect1.Before"][0].Start) {
			t.Error("Before start before Around end")
		}
		if !sum.Logs["Aspect1.Before"][0].End.Before(sum.Logs["Aspect1.Base1"][0].Start) {
			t.Error("Base1 start before Before end")
		}
		if !sum.Logs["Aspect1.Base1"][0].End.Before(sum.Logs["Aspect1.After"][0].Start) {
			t.Error("After start before Base1 end")
		}
		if !sum.Logs["Aspect1.After"][0].End.Before(sum.Logs["Aspect1.Around"][1].Start) {
			t.Error("Around start before After end")
		}
	}
}
