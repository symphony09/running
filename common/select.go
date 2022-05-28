package common

import (
	"context"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

type SelectCluster struct {
	running.Base

	Selected string

	Watch string
}

func NewSelectCluster(name string, props running.Props) (running.Node, error) {
	helper := utils.ProxyProps(props)

	node := new(SelectCluster)
	node.SetName(name)
	node.Selected = helper.SubGetString(name, "selected")
	node.Watch = helper.SubGetString(name, "watch")

	return node, nil
}

func (cluster *SelectCluster) Run(ctx context.Context) {
	helper := utils.ProxyState(cluster.State)
	var selected string

	if cluster.Watch != "" {
		selected = helper.GetString(cluster.Watch)
	}

	if selected == "" && cluster.Selected != "" {
		selected = cluster.Selected
	}

	node := cluster.SubNodesMap[cluster.Name()+"."+selected]
	if node != nil {
		node.Run(ctx)
	}
}
