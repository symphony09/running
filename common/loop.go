package common

import (
	"context"
	"sync"
	"time"

	"running"
	"running/utils"
)

type LoopCluster struct {
	running.Base

	loopCount uint

	maxLoop int

	watch string

	wait int

	wg sync.WaitGroup
}

func NewLoopCluster(name string, props running.Props) (running.Node, error) {
	helper := utils.ProxyProps(props)

	node := new(LoopCluster)
	node.SetName(name)
	node.maxLoop = helper.SubGetInt(name, "max_loop")
	node.watch = helper.SubGetString(name, "watch")
	node.wait = helper.SubGetInt(name, "wait")

	return node, nil
}

func (cluster *LoopCluster) Run(ctx context.Context) {
	helper := utils.ProxyState(cluster.State)

	for {
		if cluster.maxLoop > 0 && int(cluster.loopCount) >= cluster.maxLoop {
			break
		}

		if cluster.watch != "" {
			loop := helper.GetBool(cluster.watch)
			if !loop {
				break
			}
		}

		for _, node := range cluster.SubNodes {
			cluster.wg.Add(1)

			go func(node running.Node) {
				defer cluster.wg.Done()

				node.Run(ctx)
			}(node)
		}

		cluster.wg.Wait()
		cluster.loopCount++

		if cluster.wait > 0 {
			time.Sleep(time.Duration(cluster.wait) * time.Millisecond)
		}
	}

	cluster.loopCount = 0
}
