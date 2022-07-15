package test

import (
	"context"
	"testing"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func TestAspectCluster(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("Aspect", "Aspect1"),
		running.AddNodes("BaseTest", "Base1"),
		running.MergeNodes("Aspect1", "Base1"),
		running.SLinkNodes("Aspect1", "END"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Aspect1.around": func(point *common.JoinPoint) {
			current := time.Now()
			logName := point.Node.Name() + ".Around"
			utils.AddLog(point.State, logName, current, current, "", nil)
		},
		"Aspect1.before": func(point *common.JoinPoint) {
			current := time.Now()
			logName := point.Node.Name() + ".Before"
			utils.AddLog(point.State, logName, current, current, "", nil)
		},
		"Aspect1.after": func(point *common.JoinPoint) {
			current := time.Now()
			logName := point.Node.Name() + ".After"
			utils.AddLog(point.State, logName, current, current, "", nil)
		},
	})

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestAspectCluster", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	output := <-running.ExecPlan("TestAspectCluster", context.Background())

	if output.Err != nil {
		t.Errorf("exec plan failed,err=%s", output.Err.Error())
	} else {
		sum := utils.GetRunSummary(output.State)
		if sum.Logs["Aspect1.Base1.Around"][0].End.After(sum.Logs["Aspect1.Base1.Before"][0].Start) {
			t.Error("Before start before Around end")
		}
		if sum.Logs["Aspect1.Base1.Before"][0].End.After(sum.Logs["Aspect1.Base1"][0].Start) {
			t.Error("Base1 start before Before end")
		}
		if sum.Logs["Aspect1.Base1"][0].End.After(sum.Logs["Aspect1.Base1.After"][0].Start) {
			t.Error("After start before Base1 end")
		}
		if sum.Logs["Aspect1.Base1.After"][0].End.After(sum.Logs["Aspect1.Base1.Around"][1].Start) {
			t.Error("Around start before After end")
		}
	}
}
