package common

import (
	"context"
	"sync"
	"time"

	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

type LoopCluster struct {
	running.Base

	loopCount uint

	MaxLoop int

	Watch string

	Wait int

	wg sync.WaitGroup
}

func NewLoopCluster(name string, props running.Props) (running.Node, error) {
	helper := utils.ProxyProps(props)

	node := new(LoopCluster)
	node.SetName(name)
	node.MaxLoop = helper.SubGetInt(name, "max_loop")
	node.Watch = helper.SubGetString(name, "watch")
	node.Wait = helper.SubGetInt(name, "wait")

	return node, nil
}

func (cluster *LoopCluster) Run(ctx context.Context) {
	helper := utils.ProxyState(cluster.State)

	for {
		if cluster.MaxLoop > 0 && int(cluster.loopCount) >= cluster.MaxLoop {
			break
		}

		if cluster.Watch != "" {
			loop := helper.GetBool(cluster.Watch)
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

		if cluster.Wait > 0 {
			time.Sleep(time.Duration(cluster.Wait) * time.Millisecond)
		}
	}

	cluster.loopCount = 0
}
