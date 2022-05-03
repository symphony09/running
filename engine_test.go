package running_test

import (
	"context"
	"fmt"
	"running"
	"testing"
)

type TestNode1 struct {
	running.Base
}

func (node *TestNode1) Run(ctx context.Context) {
	fmt.Println("Node 1 running")

	//for _, n := range node.Base.SubNodes {
	//	n.Run(ctx)
	//}
}

type TestNode2 struct {
	running.Base
}

func (node *TestNode2) Run(ctx context.Context) {
	//fmt.Println("Node 2 running")

	for _, n := range node.Base.SubNodes {
		n.Run(ctx)
	}
}

type TestNode3 struct {
	running.Base
}

func (node *TestNode3) Run(ctx context.Context) {
	//fmt.Println("Node 3 running")

	for _, n := range node.Base.SubNodes {
		n.Run(ctx)
	}
}

func TestEngine(t *testing.T) {
	running.Global.RegisterNodeBuilder("A", func(props running.Props) running.Node {
		return &TestNode1{}
	})
	running.Global.RegisterNodeBuilder("B", func(props running.Props) running.Node {
		return &TestNode2{}
	})
	running.Global.RegisterNodeBuilder("C", func(props running.Props) running.Node {
		return &TestNode3{}
	})

	ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4"),
		running.AddNodes("B", "B1"),
		running.AddNodes("C", "C1"),
		running.LinkNodes("B1", "A2", "C1"),
		running.MergeNodes("A3", "A4"),
		running.MergeNodes("B1", "A1", "A3"),
		running.MergeNodes("C1", "A4"),
	}

	plan := running.NewPlan(running.EmptyProps{}, ops...)

	running.Global.RegisterPlan("P1", plan)

	out := <-running.Global.ExecPlan("P1", context.Background())

	fmt.Println(out)
}
