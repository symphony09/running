package common

import (
	"context"
	"sync"

	"github.com/symphony09/running"
)

type MergeCluster struct {
	running.Base

	HandleMerge func(state, subState running.State)

	subStates []running.State

	wg sync.WaitGroup
}

func NewMergeCluster(name string, props running.Props) (running.Node, error) {
	node := new(MergeCluster)
	node.SetName(name)

	merge, _ := props.SubGet(name, "merge")
	if method, ok := merge.(func(state, subState running.State)); ok {
		node.HandleMerge = method
	}

	return node, nil
}

func (cluster *MergeCluster) Run(ctx context.Context) {
	for i, node := range cluster.SubNodes {
		cluster.wg.Add(1)

		go func(node running.Node, i int) {
			defer cluster.wg.Done()

			node.Run(ctx)

			cluster.HandleMerge(cluster.State, cluster.subStates[i])
		}(node, i)
	}

	cluster.wg.Wait()
}

func (cluster *MergeCluster) Bind(state running.State) {
	cluster.State = state

	for _, node := range cluster.SubNodes {
		subState := NewOverlayState(state, running.NewStandardState())
		cluster.subStates = append(cluster.subStates, subState)

		if statefulNode, ok := node.(running.Stateful); ok {
			statefulNode.Bind(subState)
		}
	}
}
