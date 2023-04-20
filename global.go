package running

import "context"

var Global = NewDefaultEngine()

func NewDefaultEngine() *Engine {
	return &Engine{
		StateBuilder: func() State {
			return NewStandardState()
		},

		builders: map[string]BuildNodeFunc{},

		plans: map[string]*Plan{},

		pools: map[string]*_WorkerPool{},
	}
}

// RegisterNodeBuilder register node builder to Global
func RegisterNodeBuilder(name string, builder BuildNodeFunc) {
	Global.RegisterNodeBuilder(name, builder)
}

// RegisterPlan register plan to Global
func RegisterPlan(name string, plan *Plan) error {
	return Global.RegisterPlan(name, plan)
}

// ExecPlan exec plan register in Global
func ExecPlan(name string, ctx context.Context) <-chan Output {
	return Global.ExecPlan(name, ctx)
}

// UpdatePlan update plan register in Global.
func UpdatePlan(name string, update func(plan *Plan)) error {
	return Global.UpdatePlan(name, update)
}

// ExportPlan export plan register in Global, return json bytes
func ExportPlan(name string) ([]byte, error) {
	return Global.ExportPlan(name)
}

// WarmupPool warm up pool to avoid cold start
// name: plan name
// size: set size of worker buf queue
func WarmupPool(name string, size int) {
	Global.WarmupPool(name, size)
}

// ClearPool clear worker pool of plan, invoke it to make plan effect immediately after update
// name: name of plan
func ClearPool(name string) {
	Global.ClearPool(name)
}

// LoadPlanFromJson load plan from json data
// name: name of plan to load
// jsonData: json data of plan
// prebuilt: prebuilt nodes, can be nil
func LoadPlanFromJson(name string, jsonData []byte, prebuilt []Node) error {
	return Global.LoadPlanFromJson(name, jsonData, prebuilt)
}

// SetNodeBuilderInfo set meta info of node builder
func SetNodeBuilderInfo(name string, info NodeBuilderInfo) {
	Global.SetNodeBuilderInfo(name, info)
}
