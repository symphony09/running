package common

import "github.com/symphony09/running"

func init() {
	running.Global.RegisterNodeBuilder("Loop", NewLoopCluster)

	running.Global.RegisterNodeBuilder("Select", NewSelectCluster)

	running.Global.RegisterNodeBuilder("Serial", NewSerialCluster)

	running.Global.RegisterNodeBuilder("Aspect", NewAspectCluster)
}
