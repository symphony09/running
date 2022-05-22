package common

import "running"

func init() {
	running.Global.RegisterNodeBuilder("Loop", NewLoopCluster)

	running.Global.RegisterNodeBuilder("Select", NewSelectCluster)
}
