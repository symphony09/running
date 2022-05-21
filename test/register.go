package test

import (
	"running"
	"running/utils"
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
}
