package common

import (
	"context"
	"sync"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

type SwitchCluster struct {
	running.Base

	Status string

	Watch string

	wg sync.WaitGroup
}

func NewSwitchCluster(name string, props running.Props) (running.Node, error) {
	helper := utils.ProxyProps(props)

	node := new(SwitchCluster)
	node.SetName(name)
	node.Status = helper.SubGetString(name, "status")
	node.Watch = helper.SubGetString(name, "watch")

	return node, nil
}

func (cluster *SwitchCluster) Run(ctx context.Context) {
	helper := utils.ProxyState(cluster.State)
	var status string

	if cluster.Watch != "" {
		status = helper.GetString(cluster.Watch)
	}

	if status == "" && cluster.Status != "" {
		status = cluster.Status
	}

	if status == "on" {
		for _, node := range cluster.SubNodes {
			cluster.wg.Add(1)

			go func(node running.Node) {
				defer cluster.wg.Done()

				node.Run(ctx)
			}(node)
		}

		cluster.wg.Wait()
	}
}
