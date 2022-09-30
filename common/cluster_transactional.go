package common

import (
	"context"

	"github.com/symphony09/running"
)

type TransactionalCluster struct {
	running.Base
}

func NewTransactionalCluster(name string, props running.Props) (running.Node, error) {
	node := new(TransactionalCluster)
	node.SetName(name)
	return node, nil
}

func (cluster *TransactionalCluster) Run(ctx context.Context) {
	var count int

	defer func() {
		if r := recover(); r != nil {
			if count < len(cluster.SubNodes) {
				for i := count; i >= 0; i-- {
					if reversibleNode, ok := cluster.SubNodes[i].(running.Reversible); ok {
						reversibleNode.Revert(ctx)
					}
				}
			}

			panic(r)
		}
	}()

	for i, node := range cluster.SubNodes {
		count = i
		node.Run(ctx)
	}
}
