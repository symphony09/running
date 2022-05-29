package example

import (
	"context"
	"fmt"
	"log"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
	"github.com/symphony09/running/utils"
)

func BaseUsage() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Introduce",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("This is", ctx.Value("name"), ".")
		}))

	err := running.RegisterPlan("Plan1",
		running.NewPlan(nil, nil,
			running.AddNodes("Greet", "Greet1"),
			running.AddNodes("Introduce", "Introduce1"),
			running.SLinkNodes("Greet1", "Introduce1")))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.WithValue(context.Background(), "name", "RUNNING")

	<-running.ExecPlan("Plan1", ctx)
}

type IntroduceNode struct {
	running.Base

	Words string
}

func NewIntroduceNode(name string, props running.Props) (running.Node, error) {
	node := new(IntroduceNode)
	node.SetName(name)

	helper := utils.ProxyProps(props)
	node.Words = helper.SubGetString(name, "words")

	return node, nil
}

func (i *IntroduceNode) Run(ctx context.Context) {
	fmt.Println(i.Words)
}

func BaseUsage02() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Introduce", NewIntroduceNode)

	props := running.StandardProps(map[string]interface{}{
		"Introduce1.words": "This is RUNNING .",
	})

	err := running.RegisterPlan("Plan2",
		running.NewPlan(props, nil,
			running.AddNodes("Greet", "Greet1"),
			running.AddNodes("Introduce", "Introduce1"),
			running.SLinkNodes("Greet1", "Introduce1")))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan2", ctx)
}

func BaseUsage03() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Bye",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("bye!")
		}))

	running.RegisterNodeBuilder("Introduce", NewIntroduceNode)

	ops := []running.Option{
		running.AddNodes("Greet", "Greet1"),
		running.AddNodes("Bye", "Bye1"),
		running.AddNodes("Introduce", "Introduce1", "Introduce2", "Introduce3"),
		running.AddNodes("Select", "Select1"),
		running.MergeNodes("Select1", "Introduce2", "Introduce3"),
		running.LinkNodes("Greet1", "Select1", "Introduce1"),
		running.SLinkNodes("Introduce1", "Bye1"),
		running.SLinkNodes("Select1", "Bye1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Introduce1.words":         "This is RUNNING .",
		"Select1.Introduce2.words": "A good day .",
		"Select1.Introduce3.words": "A terrible day .",
		"Select1.selected":         "Introduce2",
	})

	err := running.RegisterPlan("Plan3", running.NewPlan(props, nil, ops...))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan3", ctx)
}

type Counter struct {
	running.Base
}

func (node *Counter) Run(ctx context.Context) {
	node.State.Transform("count", func(from interface{}) interface{} {
		if from == nil {
			return 1
		}
		if count, ok := from.(int); ok {
			count++
			return count
		} else {
			return from
		}
	})
}

type Reporter struct {
	running.Base
}

func (node *Reporter) Run(ctx context.Context) {
	count, _ := node.State.Query("count")
	fmt.Printf("count = %d\n", count)
}

func BaseUsage04() {
	running.RegisterNodeBuilder("Counter", func(name string, props running.Props) (running.Node, error) {
		node := new(Counter)
		node.SetName(name)
		return node, nil
	})

	running.RegisterNodeBuilder("Reporter", func(name string, props running.Props) (running.Node, error) {
		node := new(Reporter)
		node.SetName(name)
		return node, nil
	})

	ops := []running.Option{
		running.AddNodes("Counter", "Counter1"),
		running.AddNodes("Reporter", "Reporter1"),
		running.AddNodes("Loop", "Loop1"),
		running.MergeNodes("Loop1", "Counter1"),
		running.SLinkNodes("Loop1", "Reporter1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Loop1.max_loop": 3,
	})

	err := running.RegisterPlan("Plan4", running.NewPlan(props, nil, ops...))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan4", ctx)
}
