package test

import (
	"context"
	"testing"
	"time"

	"github.com/symphony09/running"
)

func TestAsyncWrapper(t *testing.T) {
	ops := []running.Option{
		running.AddNodes("HighCost", "C1"),
		running.WrapNodes("Async", "C1"),
		running.SLinkNodes("C1", "END"),
	}

	plan := running.NewPlan(nil, nil, ops...)

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
