package running_test

import (
	"context"
	"fmt"
	"math/rand"
	"running"
	"sync"
	"testing"
	"time"
)

type TestNode1 struct {
	running.Base
}

func (node *TestNode1) Run(ctx context.Context) {
	fmt.Printf("Single Node %s running\n", node.Base.NodeName)
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	fmt.Printf("Single Node %s stopped\n", node.Base.NodeName)
}

type TestNode2 struct {
	running.Base
}

func (node *TestNode2) Run(ctx context.Context) {
	fmt.Printf("Cluster %s running\n", node.Base.NodeName)

	for _, n := range node.Base.SubNodes {
		n.Run(ctx)
	}

	fmt.Printf("Cluster %s stopped\n", node.Base.NodeName)
}

type TestNode3 struct {
	running.Base
}

func (node *TestNode3) Run(ctx context.Context) {
	fmt.Printf("Cluster %s running\n", node.Base.NodeName)

	var wg sync.WaitGroup

	for _, n := range node.Base.SubNodes {
		wg.Add(1)

		go func(node running.Node) {
			node.Run(ctx)

			wg.Done()
		}(n)
	}

	wg.Wait()
	fmt.Printf("Cluster %s stopped\n", node.Base.NodeName)
}

type TestNode4 struct {
	running.Base
}

func (node *TestNode4) Run(ctx context.Context) {
}

type TestNode5 struct {
	running.Base
}

func (node *TestNode5) Run(ctx context.Context) {
	for _, n := range node.Base.SubNodes {
		n.Run(ctx)
	}
}

func init() {
	rand.Seed(time.Now().Unix())
}

func TestEngine(t *testing.T) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) running.Node {
		node := new(TestNode1)
		node.SetName(name)
		return node
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) running.Node {
		node := new(TestNode2)
		node.SetName(name)
		return node
	})
	running.Global.RegisterNodeBuilder("C", func(name string, props running.Props) running.Node {
		node := new(TestNode3)
		node.SetName(name)
		return node
	})

	ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4", "A5"),
		running.AddNodes("B", "B1"),
		running.AddNodes("C", "C1"),
		running.LinkNodes("B1", "A2", "C1"),
		running.MergeNodes("A3", "A4"),
		running.MergeNodes("B1", "A1", "A3"),
		running.MergeNodes("C1", "A4", "A5"),
	}

	plan := running.NewPlan(running.EmptyProps{}, ops...)

	running.Global.RegisterPlan("P1", plan)

	out := <-running.Global.ExecPlan("P1", context.Background())

	fmt.Println(out)
}

func BenchmarkEngine_ExecPlan(b *testing.B) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) running.Node {
		node := new(TestNode4)
		node.SetName(name)
		return node
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) running.Node {
		node := new(TestNode5)
		node.SetName(name)
		return node
	})

	ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4"),
		running.AddNodes("B", "B1", "B2"),
		running.LinkNodes("A1", "B1", "B2"),
		running.MergeNodes("B1", "A2", "A3"),
		running.MergeNodes("B1", "A4"),
	}

	plan := running.NewPlan(running.EmptyProps{}, ops...)

	running.Global.RegisterPlan("P1", plan)

	for i := 0; i < b.N; i++ {
		_ = <-running.Global.ExecPlan("P1", context.Background())
	}
}
