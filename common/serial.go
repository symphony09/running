package common

import (
	"context"

	"running"
)

type SerialCluster struct {
	running.Base
}

func NewSerialCluster(name string, props running.Props) (running.Node, error) {
	node := new(SerialCluster)
	node.SetName(name)

	return node, nil
}

func (cluster *SerialCluster) Run(ctx context.Context) {
	for _, node := range cluster.SubNodes {
		node.Run(ctx)
	}
}
