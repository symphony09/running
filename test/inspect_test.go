package test

import (
	"testing"

	"github.com/symphony09/running"
)

func TestInspect(t *testing.T) {
	e := running.NewDefaultEngine()
	e.RegisterNodeBuilder("Nothing", func(name string, props running.Props) (running.Node, error) {
		return &NothingNode{}, nil
	})
	e.RegisterNodeBuilder("TimeWrapper", func(name string, props running.Props) (running.Node, error) {
		node := new(TimerWrapper)
		return node, nil
	})

	ops := []running.Option{
		running.AddNodes("Nothing", "N1", "N2", "N3"),
		running.WrapNodes("TimeWrapper", "N1"),
		running.MergeNodes("N1", "N2"),
		running.ReUseNodes("N3"),
		running.SLinkNodes("N1", "N3"),
	}

	props := running.StandardProps(map[string]interface{}{
		"p":       0,
		"N1.p":    1,
		"N1.N2.p": 2,
		"N3.p":    3,
	})

	plan := running.NewPlan(props, nil, ops...)

	err := e.RegisterPlan("TestInspect", plan)
	if err != nil {
		t.Errorf("failed to register plan")
		return
	}

	inspector := running.Inspect(e)
	builders := inspector.GetNodeBuildersName()

	if len(inspector.GetNodeBuildersName()) != 2 {
		t.Errorf("wrong builders info, expect [Nothing, TimeWrapper], got %v", builders)
		return
	} else {
		checkMap := map[string]bool{}

		for _, builder := range builders {
			checkMap[builder] = true
		}

		if checkMap["Nothing"] == false {
			t.Errorf("wrong builders info, `Nothing` builder not found, got %v", builders)
			return
		}

		if checkMap["TimeWrapper"] == false {
			t.Errorf("wrong builders info, `TimeWrapper` builder not found, got %v", builders)
			return
		}
	}

	plans := inspector.GetPlansName()
	if len(plans) != 1 || plans[0] != "TestInspect" {
		t.Errorf("wrong plans info, expect [TestInspect], got %v", plans)
		return
	}

	planInfo := inspector.DescribePlan("TestInspect")
	if planInfo.Version == "" {
		t.Errorf("wrong `TestInspect` plan info, version is empty")
		return
	}

	if len(planInfo.GlobalProps) != 1 || planInfo.GlobalProps["p"] != 0 {
		t.Errorf("wrong `TestInspect` plan info, expect GlobalProps = map[p:0], got %v", planInfo.GlobalProps)
		return
	}

	edge := running.Edge{
		From: "N1",
		To:   "N3",
	}
	if len(planInfo.Edges) != 1 || planInfo.Edges[0] != edge {
		t.Errorf("wrong `TestInspect` plan info, expect Edge = [{N1 N3}], get %v", planInfo.Edges)
		return
	}

	if len(planInfo.Vertexes) != 2 {
		t.Errorf("wrong builders info, expect 2 vertexes, got %v", planInfo.Vertexes)
		return
	} else {
		checkMap := map[string]*running.VertexInfo{}

		for i, vertex := range planInfo.Vertexes {
			checkMap[vertex.VertexName] = &planInfo.Vertexes[i]
		}

		if checkMap["N1"] == nil {
			t.Errorf("wrong vertexes info, `N1` not found, got %v", builders)
			return
		} else {
			expectN1 := running.NodeInfo{
				NodeType: "Nothing",
				NodeName: "N1",
				Props: map[string]interface{}{
					"p": 1,
				},
				Wrappers: []string{
					"TimeWrapper",
				},
				SubNodes: []running.NodeInfo{
					{
						NodeType: "Nothing",
						NodeName: "N2",
						Props: map[string]interface{}{
							"p": 2,
						},
					},
				},
			}

			N1Info := checkMap["N1"].NodeInfo
			if N1Info.NodeType != "Nothing" ||
				N1Info.NodeName != "N1" ||
				N1Info.ReUse != false ||
				len(N1Info.Props) != 1 || N1Info.Props["p"] != 1 ||
				len(N1Info.Wrappers) != 1 || N1Info.Wrappers[0] != "TimeWrapper" {

				t.Errorf("wrong `N1` vertex info, expect %v, got %v", expectN1, N1Info)
				return
			}

			if len(N1Info.SubNodes) != 1 {
				t.Errorf("wrong `N1` vertex info, expect %v, got %v", expectN1, N1Info)
				return
			} else {
				N2Info := N1Info.SubNodes[0]
				if N2Info.NodeType != "Nothing" ||
					N2Info.NodeName != "N2" ||
					N2Info.ReUse != false ||
					len(N2Info.Props) != 1 || N2Info.Props["p"] != 2 ||
					len(N2Info.Wrappers) != 0 ||
					len(N2Info.SubNodes) != 0 {

					t.Errorf("wrong `N1` vertex info, expect %v, got %v", expectN1, N1Info)
					return
				}
			}
		}

		if checkMap["N3"] == nil {
			t.Errorf("wrong vertexes info, `N3` not found, got %v", builders)
			return
		} else {
			expectN3 := running.NodeInfo{
				NodeType: "Nothing",
				NodeName: "N3",
				Props: map[string]interface{}{
					"p": 3,
				},
				ReUse: true,
			}

			N3Info := checkMap["N3"].NodeInfo
			if N3Info.NodeType != "Nothing" ||
				N3Info.NodeName != "N3" ||
				N3Info.ReUse != true ||
				len(N3Info.Props) != 1 || N3Info.Props["p"] != 3 ||
				len(N3Info.Wrappers) != 0 ||
				len(N3Info.SubNodes) != 0 {

				t.Errorf("wrong `N3` vertex info, expect %v, got %v", expectN3, N3Info)
				return
			}
		}
	}
}
