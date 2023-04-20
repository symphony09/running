package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

type HelloNode struct {
	running.Base

	NodeName string `running:"name"`

	Username string `running:"prop: username"`

	msg string
}

func (node *HelloNode) Init() error {
	node.msg = fmt.Sprintf("%s: Hello, %s", node.NodeName, node.Username)
	return nil
}

func (node *HelloNode) Run(ctx context.Context) {
	utils.AddLog(node.State, node.Name(), time.Now(), time.Now(), node.msg, nil)
}

func TestNodeHelper(t *testing.T) {
	e := running.NewDefaultEngine()
	err := utils.RegisterNodes(e, &HelloNode{})
	if err != nil {
		t.Error(err)
	} else {
		builders := running.Inspect(e).GetNodeBuildersName()
		if len(builders) != 1 || builders[0] != "HelloNode" {
			t.Errorf("node builder not found, got: %v", builders)
		}

		infos := running.Inspect(e).GetNodeBuildersInfo()
		emptyInfo := running.NodeBuilderInfo{}

		if infos == nil || infos["HelloNode"] == emptyInfo {
			t.Error("node info not found")
		} else {
			info := infos["HelloNode"]

			if info.Type != running.TypeOfCluster {
				t.Errorf("wrong type info: %s", info.Type)
			}

			if info.From != "github.com/symphony09/running/test.TestNodeHelper" {
				t.Errorf("wrong from info: %s", info.From)
			}

			if info.Note != "Property Map: map[username:string]\nAuto regitered by util of runnning." {
				t.Errorf("wrong note info: %s", info.Note)
			}
		}

		ops := []running.Option{
			running.AddNodes("HelloNode", "H"),
			running.LinkNodes("H"),
		}

		props := running.StandardProps(map[string]interface{}{
			"H.username": "Oliver",
		})

		plan := running.NewPlan(props, nil, ops...)
		if err := e.RegisterPlan("TestNodeHelper", plan); err != nil {
			t.Error(err)
		}

		out := <-e.ExecPlan("TestNodeHelper", context.Background())
		if out.Err != nil {
			t.Error(out.Err)
		} else {
			logs := utils.GetRunSummary(out.State).Logs
			if logs["H"] == nil {
				t.Error("Node H log not found")
				return
			}

			if len(logs["H"]) != 1 {
				t.Errorf("wrong node H logs, logs = %v", logs["H"])
				return
			}

			if logs["H"][0].Msg != "H: Hello, Oliver" {
				t.Errorf("wrong node H msg, msg = %v", logs["H"][0].Msg)
				return
			}
		}
	}
}
