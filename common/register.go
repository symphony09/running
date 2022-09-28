// Package common implement some node, cluster, state for common usage
package common

import "github.com/symphony09/running"

func init() {
	running.RegisterNodeBuilder("Loop", NewLoopCluster)

	running.RegisterNodeBuilder("Select", NewSelectCluster)

	running.RegisterNodeBuilder("Serial", NewSerialCluster)

	running.RegisterNodeBuilder("Aspect", NewAspectCluster)

	running.RegisterNodeBuilder("Merge", NewMergeCluster)

	running.RegisterNodeBuilder("Switch", NewSwitchCluster)

	running.RegisterNodeBuilder("Async", NewAsyncWrapper)
}
