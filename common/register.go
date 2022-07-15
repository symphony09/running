// Package common implement some node, cluster, state for common usage
package common

import "github.com/symphony09/running"

func init() {
	running.Global.RegisterNodeBuilder("Loop", NewLoopCluster)

	running.Global.RegisterNodeBuilder("Select", NewSelectCluster)

	running.Global.RegisterNodeBuilder("Serial", NewSerialCluster)

	running.Global.RegisterNodeBuilder("Aspect", NewAspectCluster)

	running.Global.RegisterNodeBuilder("Merge", NewMergeCluster)

	running.Global.RegisterNodeBuilder("Switch", NewSwitchCluster)
}
