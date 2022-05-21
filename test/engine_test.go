package test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"running"
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
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode1)
		node.SetName(name)
		return node, nil
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode2)
		node.SetName(name)
		return node, nil
	})
	running.Global.RegisterNodeBuilder("C", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode3)
		node.SetName(name)
		return node, nil
	})

	ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4", "A5"),
		running.AddNodes("B", "B1"),
		running.AddNodes("C", "C1", "C2"),
		running.LinkNodes("B1", "A2", "C1", "C2"),
		running.MergeNodes("A3", "A4"),
		running.MergeNodes("B1", "A1", "A3"),
		running.MergeNodes("C1", "A4", "A5"),
	}

	c2 := new(TestNode3)
	c2.SetName("C2")
	a6 := new(TestNode2)
	a6.SetName("C2.A6")
	c2.Inject([]running.Node{a6})

	plan := running.NewPlan(running.EmptyProps{}, []running.Node{c2}, ops...)

	err := running.Global.RegisterPlan("P1", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("P1", context.Background())

	fmt.Println(out)
}

type TestNode6 struct {
	running.Base

	chosen string
}

func (node *TestNode6) Run(ctx context.Context) {
	fmt.Printf("Cluster %s running\n", node.Name())

	for _, subNode := range node.Base.SubNodes {
		if subNode.Name() == node.chosen {
			subNode.Run(ctx)
		}
	}

	fmt.Printf("Cluster %s stopped\n", node.Name())
}

func TestProps(t *testing.T) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode6)
		node.SetName(name)
		chosen, _ := props.Get(name + ".chosen")
		node.chosen, _ = chosen.(string)
		node.chosen = name + "." + node.chosen
		return node, nil
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode1)
		node.SetName(name)
		return node, nil
	})

	props := running.StandardProps(map[string]interface{}{"A1.chosen": "B2"})

	ops := []running.Option{
		running.AddNodes("A", "A1"),
		running.AddNodes("B", "B1", "B2", "B3"),
		running.MergeNodes("A1", "B1", "B2", "B3"),
		running.LinkNodes("A1"),
	}

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("P2", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("P2", context.Background())

	fmt.Println(out)
}

func TestEngine_UpdatePlan(t *testing.T) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode6)
		node.SetName(name)
		chosen, _ := props.Get(name + ".chosen")
		node.chosen, _ = chosen.(string)
		node.chosen = name + "." + node.chosen
		return node, nil
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode1)
		node.SetName(name)
		return node, nil
	})

	props := running.StandardProps(map[string]interface{}{"A1.chosen": "B2"})

	ops := []running.Option{
		running.AddNodes("A", "A1"),
		running.AddNodes("B", "B1", "B2", "B3"),
		running.MergeNodes("A1", "B1", "B2", "B3"),
		running.LinkNodes("A1"),
	}

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("P2", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	out := <-running.Global.ExecPlan("P2", context.Background())

	fmt.Println(out)

	err = running.Global.UpdatePlan("P2", true, func(plan *running.Plan) {
		plan.Props = running.StandardProps(map[string]interface{}{"A1.chosen": "B3"})
	})
	if err != nil {
		t.Errorf("failed to update plan")
		return
	}

	out = <-running.Global.ExecPlan("P2", context.Background())

	fmt.Println(out)
}

type TestNode7 struct {
	running.Base
}

func (node *TestNode7) Run(ctx context.Context) {
	fmt.Printf("Single Node %s running\n", node.Name())

	select {
	case <-time.After(150 * time.Millisecond):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}

	fmt.Printf("Single Node %s stopped\n", node.Name())
}

type TestNode8 struct {
	running.Base

	loop int
}

func (node *TestNode8) Run(ctx context.Context) {
	fmt.Printf("Cluster %s running\n", node.Name())

	for i := 0; i < node.loop; i++ {
		for _, subNode := range node.Base.SubNodes {
			subNode.Run(ctx)
		}
	}

	fmt.Printf("Cluster %s stopped\n", node.Name())
}

func TestCtx(t *testing.T) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode8)
		node.SetName(name)
		loop, _ := props.Get(name + ".loop")
		node.loop, _ = loop.(int)
		return node, nil
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode7)
		node.SetName(name)
		return node, nil
	})

	props := running.StandardProps(map[string]interface{}{"A1.loop": 5})

	ops := []running.Option{
		running.AddNodes("A", "A1"),
		running.AddNodes("B", "B1"),
		running.MergeNodes("A1", "B1"),
		running.LinkNodes("A1"),
	}

	plan := running.NewPlan(props, nil, ops...)

	err := running.Global.RegisterPlan("P3", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	out := <-running.Global.ExecPlan("P3", ctx)

	fmt.Println(out)
}

func BenchmarkEngine_ExecPlan(b *testing.B) {
	running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode4)
		node.SetName(name)
		return node, nil
	})
	running.Global.RegisterNodeBuilder("B", func(name string, props running.Props) (running.Node, error) {
		node := new(TestNode5)
		node.SetName(name)
		return node, nil
	})

	ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4"),
		running.AddNodes("B", "B1", "B2"),
		running.LinkNodes("A1", "B1", "B2"),
		running.MergeNodes("B1", "A2", "A3"),
		running.MergeNodes("B1", "A4"),
	}

	plan := running.NewPlan(running.EmptyProps{}, nil, ops...)

	err := running.Global.RegisterPlan("P1", plan)
	if err != nil {
		b.Errorf("failed to register plan")
		return
	}

	for i := 0; i < b.N; i++ {
		_ = <-running.Global.ExecPlan("P1", context.Background())
	}
}
