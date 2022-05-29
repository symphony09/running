# Running

## Running 是什么

Running 是一个基于 DAG 的 Golang 图化执行框架。

目标是实现方便，灵活地切换算子的组合方式和执行顺序，并发挥 Golang 的并发优势。

### 特点

- 定义 Node，定义 plan，执行 plan 三步走
- 内置基本实现，目标开箱即用
- 可以并行执行的就并行执行，目标高性能
- 无任何第三方依赖，目标稳定可靠

## 使用说明

### 简单使用

#### 示例代码

```go
package example

import (
	"context"
	"fmt"
	"log"

	"github.com/symphony09/running"
	"github.com/symphony09/running/common"
)

func BaseUsage() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Introduce",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("This is", ctx.Value("name"), ".")
		}))

	err := running.RegisterPlan("Plan1",
		running.NewPlan(nil, nil,
			running.AddNodes("Greet", "Greet1"),
			running.AddNodes("Introduce", "Introduce1"),
			running.SLinkNodes("Greet1", "Introduce1")))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.WithValue(context.Background(), "name", "RUNNING")

	<-running.ExecPlan("Plan1", ctx)
}

```

#### 输出

```
Hello!
This is RUNNING .
```

#### 说明

示例代码做了以下几件事：

1. 注册 Greet，Introduce 两个 Node 构建函数

**Node 是引擎的执行单位**，引擎会管理 Node 的构建和执行。所以需要注册 Node 的构建函数，而不是具体的 Node。

`common.NewSimpleNodeBuilder` 接受一个签名为 `func(ctx context.Context)` 的函数，返回 SimpleNode 的构建函数。

SimpleNode 是引擎的一个内置 Node 实现。


NewSimpleNodeBuilder 接受的函数会封装在 SimpleNode 内，引擎执行 SimpleNode  时就会调用此函数。

2. 注册 Plan1

**Plan 是引擎的执行规划**，有了封装了运算逻辑的 Node 后，就可以规划如何执行 Node 了。

这里先忽略`running.NewPlan` 的前两个参数，第三个参数开始是不定长参数，定义了一系列操作：

- AddNodes：添加 Node

  - 第一个参数是 Node 类型，对应之前注册的 Node 构建函数
  - 第二个参数开始是不定长参数，对应具体 Node 的名字。几个名字，就对应几个 Node。
- SLinkNodes：不定长参数，将添加的 Node 串行连接起来

示例代码的 plan 可以简单表示为 Greet1 -> Introduce1

Greet1 由 Greet 对应的构建函数构建，执行时输出 Hello!

Introduce1 由 Introduce 对应的构建函数构建，执行时输出 This is 加上上下文参数中的 name 值。

3. 执行 Plan1

在 plan 注册完成后就可以在任意时机，执行任意次数 plan。

`running.ExecPlan` 接受两个参数。一个是 Plan 名，另一个是执行的上下文参数，上下文参数会由引擎传递给 Node 的运行函数。

ExecPlan 会立即返回一个通道，真正的执行逻辑是异步执行的，最后将结果通过通道返回。

### 自定义 Node

当引擎内置的 Node 实现不能满足需要时，可以自定义 Node 使用。

Node 接口定义：

```go
type Node interface {
	Name() string

	Run(ctx context.Context)

	Reset()
}
```

Node 接口共三个方法，分别用于获取 Node 名，执行运行逻辑，重置 Node 状态。

重置方法在当次计划执行过程中 Node 不会再执行时调用。引擎不会每次都创建新的 Node 来执行计划，所以需要通过重置方法来初始化 Node。

#### 示例代码

```go
type IntroduceNode struct {
	running.Base

	Words string
}

func NewIntroduceNode(name string, props running.Props) (running.Node, error) {
	node := new(IntroduceNode)
	node.SetName(name)

	helper := utils.ProxyProps(props)
	node.Words = helper.SubGetString(name, "words")

	return node, nil
}

func (i *IntroduceNode) Run(ctx context.Context) {
	fmt.Println(i.Words)
}

func BaseUsage02() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Introduce", NewIntroduceNode)

	props := running.StandardProps(map[string]interface{}{
		"Introduce1.words": "This is RUNNING .",
	})

	err := running.RegisterPlan("Plan2",
		running.NewPlan(props, nil,
			running.AddNodes("Greet", "Greet1"),
			running.AddNodes("Introduce", "Introduce1"),
			running.SLinkNodes("Greet1", "Introduce1")))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan2", ctx)
}
```

#### 输出

```go
Hello!
This is RUNNING .
```

#### 说明

有几点需要说明：

- IntroduceNode 不需要实现 Name 和 Reset 是因为嵌入的 running.Base 已经实现了，SetName 也是 running.Base 实现的。
- Props 用于为 Node 构建函数提供构建参数
  - 引擎内置了一个基于 Map 的 Props 实现，即 StandardProps，Key 格式为 Node 名 + “.” + 参数名
  - utils.ProxyProps 用于简化参数类型断言

### 更复杂的 Plan

上文提到了 AddNodes 和 SLinkNodes，除了这两种引擎还支持 MergeNodes 和 LinkNodes。

- MergeNodes ：将一些 Node 合并为 一个 Node 的 子 Node，子 Node 如何执行由父 Node 决定。
- LinkNodes：与 SLinkNodes 类似，但连接方式略有不同，LinkNodes 是将其他 Node 同时作为一个 Node的后继。

#### 示例代码

```go
func BaseUsage03() {
	running.RegisterNodeBuilder("Greet",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("Hello!")
		}))

	running.RegisterNodeBuilder("Bye",
		common.NewSimpleNodeBuilder(func(ctx context.Context) {
			fmt.Println("bye!")
		}))

	running.RegisterNodeBuilder("Introduce", NewIntroduceNode)

	ops := []running.Option{
		running.AddNodes("Greet", "Greet1"),
		running.AddNodes("Bye", "Bye1"),
		running.AddNodes("Introduce", "Introduce1", "Introduce2", "Introduce3"),
		running.AddNodes("Select", "Select1"),
		running.MergeNodes("Select1", "Introduce2", "Introduce3"),
		running.LinkNodes("Greet1", "Select1", "Introduce1"),
		running.SLinkNodes("Introduce1", "Bye1"),
		running.SLinkNodes("Select1", "Bye1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Introduce1.words":         "This is RUNNING .",
		"Select1.Introduce2.words": "A good day .",
		"Select1.Introduce3.words": "A terrible day .",
		"Select1.selected":         "Introduce2",
	})

	err := running.RegisterPlan("Plan3", running.NewPlan(props, nil, ops...))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan3", ctx)
}
```

#### 输出

```
Hello!
This is RUNNING .
A good day .
bye!
```

#### 说明

程序执行流程如下：

Greet1 -> Introduce1 ->  Select1.Introduce2 -> Bye1

或

Greet1 ->  Select1.Introduce2 -> Introduce1 -> Bye1

示例代码中，Select 是在引入 common 包时自动注册的 Node，Select  可以合并其他 Node，称为 Cluster （簇）。

Select 会根据 props 传入的参数，从合并的 Node 中选择 Node 执行。

### 自定义 Cluster

Cluster 接口定义：

```go
type Cluster interface {
	Node

	Inject(nodes []Node)
}
```

Inject 方法用于引擎根据 plan 注入子 Node，嵌入 running.Base 可以自动实现此方法

```go
func (base *Base) Inject(nodes []Node) {
	base.SubNodes = append(base.SubNodes, nodes...)

	if base.SubNodesMap == nil {
		base.SubNodesMap = make(map[string]Node)
	}

	for _, node := range nodes {
		base.SubNodesMap[node.Name()] = node
	}
}
```

嵌入 running.Base 的结构体可以间接通过 Base 获取 SubNodes 和 SubNodesMap 字段，从而执行这些 Node。

具体实现方法可以参考 common 包下的源码。

### Node 间通信

在引擎中，Node 间通过 State 通信，执行完成后 State 也会作为 ExecPlan 的执行结果从通道返回。

要使用 State，Node 需要实现 Stateful 接口：

```go
type Stateful interface {
	Node

	Bind(state State)
}
```

Bind 用于引擎为 Node 绑定状态，嵌入 running.Base 可以自动实现此方法。

```go
func (base *Base) Bind(state State) {
	base.State = state

	for _, node := range base.SubNodes {
		if statefulNode, ok := node.(Stateful); ok {
			statefulNode.Bind(state)
		}
	}
}
```

嵌入 running.Base 的结构体可以间接通过 Base 获取 State 字段，从而读取和写入 State。

State 定义如下：

```go
type State interface {
	Query(key string) (interface{}, bool)

	Update(key string, value interface{})

	Transform(key string, transform TransformStateFunc)
}
```

分别用于查询 State，更新 State 和转换 State，引擎内置了 并发安全的 StandardState 实现。

#### 示例代码

```go
type Counter struct {
	running.Base
}

func (node *Counter) Run(ctx context.Context) {
	node.State.Transform("count", func(from interface{}) interface{} {
		if from == nil {
			return 1
		}
		if count, ok := from.(int); ok {
			count++
			return count
		} else {
			return from
		}
	})
}

type Reporter struct {
	running.Base
}

func (node *Reporter) Run(ctx context.Context) {
	count, _ := node.State.Query("count")
	fmt.Printf("count = %d\n", count)
}

func BaseUsage04() {
	running.RegisterNodeBuilder("Counter", func(name string, props running.Props) (running.Node, error) {
		node := new(Counter)
		node.SetName(name)
		return node, nil
	})

	running.RegisterNodeBuilder("Reporter", func(name string, props running.Props) (running.Node, error) {
		node := new(Reporter)
		node.SetName(name)
		return node, nil
	})

	ops := []running.Option{
		running.AddNodes("Counter", "Counter1"),
		running.AddNodes("Reporter", "Reporter1"),
		running.AddNodes("Loop", "Loop1"),
		running.MergeNodes("Loop1", "Counter1"),
		running.SLinkNodes("Loop1", "Reporter1"),
	}

	props := running.StandardProps(map[string]interface{}{
		"Loop1.max_loop": 3,
	})

	err := running.RegisterPlan("Plan4", running.NewPlan(props, nil, ops...))

	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	<-running.ExecPlan("Plan4", ctx)
}
```

#### 输出

```
count = 3
```

#### 说明

Loop 也是 common 包内定义的 cluster，可以按 props 中的 max_loop 参数循环执行子Node。

Loop1 的 max_loop 参数设为 3，则 Counter1 最多循环执行 3 次。

Counter1 执行时将 count 写入 State，而 Reporter1 执行时从 State 读取 count 并打印。

utils 包也有简化 State 类型断言的 helper。

ExecPlan 返回的通道的基础类型为 <-chan Output，Output 定义如下：

```go
type Output struct {
	Err error

	State State
}
```

如果 plan 顺利执行，引擎会把 State 透出供外部代码使用。

### 更新 Plan

更新函数为`running.UdatePlan`，签名为 `func UpdatePlan(name string, fastMode bool, update func(plan *Plan)) error`

第一个参数为要更新的 plan 名，第二个参数设置是否快速生效，如果设为 true 即快速生效，Worker 池会被清空，第三个参数则是 plan 的具体更新函数。

### 预建 Node

有时，构建 Node 的成本是高昂的，虽然引擎已经通过 Worker 池复用 Node 来减小开销，

但是在需要新建 Worker 的情况下，还是会存在开销过大的问题，这在 plan 执行次数还比较少时或突然加快执行频率时会比较突出。

为了解决这个问题，引擎支持从 plan 中获取预先构建好的 Node 的复制而不是重新构建 Node 来减小构建开销。

`running.NewPlan` 的第二个参数用于接收预建 Node 数组。如：

```
c2 := new(TestNode3)
c2.SetName("C2")
a6 := new(TestNode2)
a6.SetName("C2.A6")
c2.Inject([]running.Node{a6})

plan := running.NewPlan(running.EmptyProps{}, []running.Node{c2}, ops...)
```

要注意的是，多个复制而来的 Node 之间可能通过指针相互影响，这通常不是我们所期望的。

所以最好为预建 Node 实现 Cloneable 接口：

```go
type Cloneable interface {
	Node

	Clone() Node
}
```

这样引擎就会调用预建 Node 的 Clone 方法获取克隆Node，而不是直接浅拷贝预建 Node。
