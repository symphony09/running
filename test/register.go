package test

import (
	"github.com/symphony09/running"
	"github.com/symphony09/running/utils"
)

func init() {
	running.Global.RegisterNodeBuilder("BaseTest", func(name string, props running.Props) (running.Node, error) {
		node := new(BaseTestNode)
		node.SetName(name)
		return node, nil
	})

	running.Global.RegisterNodeBuilder("SetState", func(name string, props running.Props) (running.Node, error) {
		node := new(SetStateNode)
		node.SetName(name)

		helper := utils.ProxyProps(props)
		node.key = helper.SubGetString(name, "key")
		node.value, _ = props.SubGet(name, "value")

		return node, nil
	})

	running.RegisterNodeBuilder("Nothing", func(name string, props running.Props) (running.Node, error) {
		node := new(NothingNode)
		node.SetName(name)

		return node, nil
	})

	running.RegisterNodeBuilder("TimerWrapper", func(name string, props running.Props) (running.Node, error) {
		node := new(TimerWrapper)
		return node, nil
	})
}
