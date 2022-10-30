package running

import "strings"

type Inspector struct {
	target *Engine
}

func Inspect(e *Engine) Inspector {
	return Inspector{target: e}
}

func (i Inspector) GetNodeBuildersName() []string {
	var names []string
	if i.target != nil {
		i.target.buildersLocker.RLock()
		defer i.target.buildersLocker.RUnlock()

		for name := range i.target.builders {
			names = append(names, name)
		}
	}

	return names
}

func (i Inspector) GetPlansName() []string {
	var names []string
	if i.target != nil {
		i.target.plansLocker.RLock()
		defer i.target.plansLocker.RUnlock()

		for name := range i.target.plans {
			names = append(names, name)
		}
	}

	return names
}

type PlanInfo struct {
	Version string

	Vertexes []VertexInfo

	Edges []Edge

	GlobalProps map[string]interface{}
}

type VertexInfo struct {
	VertexName string

	NodeInfo NodeInfo
}

type NodeInfo struct {
	NodeType string

	NodeName string

	Props map[string]interface{}

	Wrappers []string

	ReUse bool

	SubNodes []NodeInfo
}

type Edge struct {
	From string
	To   string
}

func (i Inspector) DescribePlan(name string) PlanInfo {
	var info PlanInfo
	if i.target != nil {
		i.target.plansLocker.RLock()
		plan := i.target.plans[name]
		i.target.plansLocker.RUnlock()

		if plan != nil {
			plan.locker.RLock()
			defer plan.locker.RUnlock()

			info.Version = plan.version
			info.Vertexes = make([]VertexInfo, 0, len(plan.graph.Vertexes))
			for vName, vertex := range plan.graph.Vertexes {
				info.Vertexes = append(info.Vertexes, VertexInfo{
					VertexName: vName,
					NodeInfo:   describeNode(plan, vertex.RefRoot.NodeName),
				})

				for _, vNext := range vertex.Next {
					if vNext != nil && vNext.RefRoot != nil {
						info.Edges = append(info.Edges, Edge{
							From: vName,
							To:   vNext.RefRoot.NodeName,
						})
					}
				}
			}

			info.GlobalProps = map[string]interface{}{}
			if exportable, ok := plan.props.(ExportableProps); ok {
				raw := exportable.Raw()
				for k, v := range raw {
					if !strings.Contains(k, ".") {
						info.GlobalProps[k] = v
					}
				}
			}
		}
	}

	return info
}

func describeNode(plan *Plan, node string, path ...string) NodeInfo {
	ref := plan.graph.NodeRefs[node]
	info := NodeInfo{
		NodeType: ref.NodeType,
		NodeName: ref.NodeName,
		Wrappers: ref.Wrappers,
		ReUse:    ref.ReUse,

		Props:    map[string]interface{}{},
		SubNodes: make([]NodeInfo, 0, len(ref.SubRefs)),
	}

	path = append(path, node)
	prefix := strings.Join(path, ".")

	if exportable, ok := plan.props.(ExportableProps); ok {
		raw := exportable.Raw()
		for k, v := range raw {
			p := strings.LastIndex(k, ".")
			if p > 0 && p+1 < len(k) && k[:p] == prefix {
				info.Props[k[p+1:]] = v
			}
		}
	}

	for _, subRef := range ref.SubRefs {
		info.SubNodes = append(info.SubNodes, describeNode(plan, subRef.NodeName, path...))
	}

	return info
}
