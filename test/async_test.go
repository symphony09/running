package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
)

func TestAsyncWrapper(t *testing.T) {
	running.RegisterNodeBuilder("Boom",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			panic("Boom!")
		}))

	ops := []running.Option{
		running.AddNodes("HighCost", "C1"),
		running.AddNodes("Boom", "B1"),
		running.WrapNodes("Async", "C1", "B1"),
		running.SLinkNodes("C1", "B1", "END"),
	}

	props := running.StandardProps{
		"B1.panic_handler": func(ctx context.Context, nodeName string, v interface{}) {
			fmt.Printf("%s recover from %v\n", nodeName, v)
		},
	}

	plan := running.NewPlan(props, nil, ops...)

	err := running.RegisterPlan("TestAsyncWrapper", plan)
	if err != nil {
		t.Errorf("register plan failed, err=%s", err.Error())
		return
	}

	startTime := time.Now()
	output := <-running.ExecPlan("TestAsyncWrapper", context.Background())
	if output.Err != nil {
		t.Errorf("exec plan failed, err=%s\n", output.Err.Error())
	} else if time.Since(startTime).Milliseconds() > 1000 {
		t.Errorf("async job cost too much time, cost=%d ms\n", time.Since(startTime).Milliseconds())
	}
}
