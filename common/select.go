package common

import (
	"context"

	"running"
	"running/utils"
)

type SelectCluster struct {
	running.Base

	selected string

	watch string
}

func NewSelectCluster(name string, props running.Props) (running.Node, error) {
	helper := utils.ProxyProps(props)

	node := new(SelectCluster)
	node.SetName(name)
	node.selected = helper.SubGetString(name, "selected")
	node.watch = helper.SubGetString(name, "watch")

	return node, nil
}

func (cluster *SelectCluster) Run(ctx context.Context) {
	helper := utils.ProxyState(cluster.State)
	var selected string

	if cluster.watch != "" {
		selected = helper.GetString(cluster.watch)
	}

	if selected == "" && cluster.selected != "" {
		selected = cluster.selected
	}

	node := cluster.SubNodesMap[cluster.Name()+"."+selected]
	if node != nil {
		node.Run(ctx)
		node.Reset()
	}
}
