package common

import (
	"context"
	"sync"

	"github.com/symphony09/running"
)

type AspectCluster struct {
	running.Base

	Around func(cluster *AspectCluster)

	Before func(cluster *AspectCluster)

	After func(cluster *AspectCluster)

	wg sync.WaitGroup
}

func NewAspectCluster(name string, props running.Props) (running.Node, error) {
	node := new(AspectCluster)
	node.SetName(name)

	around, _ := props.SubGet(name, "around")
	if method, ok := around.(func(cluster *AspectCluster)); ok {
		node.Around = method
	}
	before, _ := props.SubGet(name, "before")
	if method, ok := before.(func(cluster *AspectCluster)); ok {
		node.Before = method
	}
	after, _ := props.SubGet(name, "after")
	if method, ok := after.(func(cluster *AspectCluster)); ok {
		node.After = method
	}

	return node, nil
}

func (cluster *AspectCluster) Run(ctx context.Context) {
	if cluster.Around != nil {
		cluster.Around(cluster)
	}

	if cluster.Before != nil {
		cluster.Before(cluster)
	}

	for _, node := range cluster.SubNodes {
		cluster.wg.Add(1)

		go func(node running.Node) {
			defer cluster.wg.Done()

			node.Run(ctx)
		}(node)
	}

	cluster.wg.Wait()

	if cluster.After != nil {
		cluster.After(cluster)
	}

	if cluster.Around != nil {
		cluster.Around(cluster)
	}
}
