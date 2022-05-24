package common

import (
	"context"
	"sync"

	"github.com/symphony09/running"
)

type AspectCluster struct {
	running.Base

	HandleAround func(point *JoinPoint)

	HandleBefore func(point *JoinPoint)

	HandleAfter func(point *JoinPoint)

	wg sync.WaitGroup
}

type JoinPoint struct {
	Ctx context.Context

	State running.State

	Node running.Node
}

func NewAspectCluster(name string, props running.Props) (running.Node, error) {
	node := new(AspectCluster)
	node.SetName(name)

	around, _ := props.SubGet(name, "around")
	if method, ok := around.(func(point *JoinPoint)); ok {
		node.HandleAround = method
	}
	before, _ := props.SubGet(name, "before")
	if method, ok := before.(func(point *JoinPoint)); ok {
		node.HandleBefore = method
	}
	after, _ := props.SubGet(name, "after")
	if method, ok := after.(func(point *JoinPoint)); ok {
		node.HandleAfter = method
	}

	return node, nil
}

func (cluster *AspectCluster) Run(ctx context.Context) {
	for _, node := range cluster.SubNodes {
		cluster.wg.Add(1)

		go func(node running.Node) {
			defer cluster.wg.Done()

			point := &JoinPoint{
				Ctx:   ctx,
				State: cluster.State,
				Node:  node,
			}

			if cluster.HandleAround != nil {
				cluster.HandleAround(point)
			}

			if cluster.HandleBefore != nil {
				cluster.HandleBefore(point)
			}

			node.Run(ctx)

			if cluster.HandleAfter != nil {
				cluster.HandleAfter(point)
			}

			if cluster.HandleAround != nil {
				cluster.HandleAround(point)
			}
		}(node)
	}

	cluster.wg.Wait()
}
