# Running

## What is running

Running 是一个基于 DAG 的 golang 图化执行框架。

目标是实现方便，灵活地切换算子的组合方式和执行顺序，并发挥 golang 的并发优势。

## How to start

大体可以分为 5 个步骤：

1. 定义节点（define node） 
2. 注册节点（register node）
3. 定义计划（define plan）
4. 注册计划（register plan） 
5. 执行计划（execute plan）

### Example

test/engine_test.go 展示了 running  的基本用法，感兴趣的话可以查看源码和尝试执行。

下面以 engine_test.go 中代码为例说明 5 个步骤如何完成。

#### 1. 定义节点示例

```go
type TestNode1 struct {
	running.Base
}

func (node *TestNode1) Run(ctx context.Context) {
	fmt.Printf("Single Node %s running\n", node.Base.NodeName)
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	fmt.Printf("Single Node %s stopped\n", node.Base.NodeName)
}
```

这是最简单的一个 node 定义，

它的功能是模拟一个 5s 内的工作负载，并在开始和结束打印提示信息。

#### 2. 注册节点示例

```go
running.Global.RegisterNodeBuilder("A", func(name string, props running.Props) (running.Node, error) {
	node := new(TestNode1)
	node.SetName(name)
	return node, nil
})
```

这样就可以在 running 中通过 A 这个名字引用 TestNode1

**注意**：注册的是节点的构建函数，而非节点本身

#### 3. 定义计划示例

```go
ops := []running.Option{
		running.AddNodes("A", "A1", "A2", "A3", "A4", "A5"),
		running.AddNodes("B", "B1"),
		running.AddNodes("C", "C1", "C2"),
		running.LinkNodes("B1", "A2", "C1", "C2"),
		running.MergeNodes("A3", "A4"),
		running.MergeNodes("B1", "A1", "A3"),
		running.MergeNodes("C1", "A4", "A5"),
	}

c2 := new(TestNode3)
c2.SetName("C2")
a6 := new(TestNode2)
a6.SetName("C2.A6")
c2.Inject([]running.Node{a6})

plan := running.NewPlan(running.EmptyProps{}, []running.Node{c2}, ops...)
```

计划（Plan）由节点初始化属性（Props），预建节点（Prebuilt）和节点操作（Options）组成。

**Props** 存放初始化节点所需要的属性

running 提供了几类基本实现，如示例中的 EmptyProps，而 StandardProps 则是更实用的实现。

这里不作展开，详细说明见下文核心概念 Props 部分。

**Prebuilt** 存放预先创建好的节点

如果节点创建比较耗时，可以提供预建节点，引擎会复用这些预先创建好的节点。

在预建节点实现了 Clone 方法的情况下，引擎会优先使用预建节点的克隆，否则直接使用节点。

**Option** 存放定义了一组操作，决定了节点的组织方式，总共有四种

- AddNodes

表示添加节点，第一个参数表示节点类型，其后则为实例命名

如`running.AddNodes("A", "A1", "A2", "A3", "A4", "A5")`

这行代码表示添加 5 个 A 类型节点，分别命名为 "A1", "A2", "A3", "A4", "A5"。

- LinkNodes

表示连接节点，第一个参数表示的节点连接其后所有节点，用于表达节点间的依赖关系

如 `running.LinkNodes("B1", "A2", "C1")`表示 B1 同时连接 A2，C1。

- SLinkNodes

与 LinkNodes 相似，但是节点之间是从头到尾串行连接

这表达了 A2，C1两个节点依赖 B1 节点，只有 B1 运行完成后 A2，C1才能开始运行。

- MergeNodes

表示合并节点，第二个参数开始代表的节点会合并为第一个的子节点

如`running.MergeNodes("B1", "A1", "A3")`，表示将 A1，A3 作为 B1 的子节点

B1 在运行时可以获取到 A1，A3 节点并决定是否及何时运行他们

**注意**：

- 通个节点实例可以被多次合并，在这种情况下它可能会被不同节点调用运行多次
- 被连接的实例也可以被合并，在这种情况下它可能会被运行多次
  - 前置依赖完成后被直接调用运行
  - 被节点调用运行

#### 4. 注册计划示例

```go
running.Global.RegisterPlan("P1", plan)
```

与注册节点类似，表示把定义的 plan 注册为 P1，之后就可以在 running 中通过 P1 这个名字引用定义好的 plan。

#### 5. 执行计划示例

```go
out := <-running.Global.ExecPlan("P1", context.Background())

fmt.Println(out)
```

`ExecPlan`传入两个参数，一个是计划名，另一个会作为节点运行的上下文信息传递给 Run 方法。

此方法返回一个 Chan 通道，用于返回最终的执行状态数据。这个执行状态数据是什么后面再讲。

#### 6× 更新计划示例

```go
err = running.Global.UpdatePlan("P2", true, func(plan *running.Plan) {
		plan.Props = running.StandardProps(map[string]interface{}{"A1.chosen": "B3"})
})
```

第一个参数表示 plan 名，第二个表示需要立即生效（这会清空 worker 池，可能造成短时间内负载快速升高）

第三个参数则是更新函数，获取原 plan 并进行更新

更新 plan 可能会由于节点存在死环而失败，这不会影响 plan 按原来方式执行，但是 plan 更新的字段不会回退。

## More

### 核心概念

#### Node

Node 是对运算逻辑的封装，即算子。Node 可以包含其他 Node，在 running 中称为 Cluster（簇）。

从示例中可以知道，算子的运行时机有两种：

一种是 running 引擎判断依赖解决后调用，一种是被作为簇的一部分调用。

第一种简单通用，第二种则可以定制更复杂的运行逻辑。

如循环执行某个节点，或根据条件选取某个节点执行，这样的执行逻辑用依赖关系表达是比较困难的。

相关接口定义如下：

```go
type Node interface {
	Name() string

	Run(ctx context.Context)

	Reset()
}

type Cluster interface {
	Node

	Inject(nodes []Node)
}
```

Run 在 Node 执行时调用，Reset 在 Node 不会再被执行时调用

由于引擎并不会在每次执行计划时都重新创建 Node，所以需要保证本次 Node 调用完成后通过重置回到初始状态。

为 Node 实现 Inject 方法，引擎就会根据 plan 为 Node 注入子 Node。怎么使用这些 Node 是 Cluster 需要解决的问题。

#### Props

props 用于提供 Node 的初始化参数，通过 plan 传递给引擎，引擎再传递给 Node 的构造函数。

相关定义如下：

```go
type Props interface {
	Get(key string) (interface{}, bool)

	SubGet(sub, key string) (interface{}, bool)
}

type BuildNodeFunc func(name string, props Props) (Node, error)
```

以 引擎提供的 StandardProps 实现为例，使用方法如下：

```go
// 约定 A1 节点的 chosen 属性为 B1，使 A1 选择子节点 B1
props := running.StandardProps(map[string]interface{}{"A1.chosen": "B2"})

// A1 节点的构造函数读取 props，进行初始化
func(name string, props running.Props) (running.Node, error) {
	node := new(TestNode6)
	node.SetName(name)
	chosen, _ := props.Get(name + ".chosen")
	node.chosen, _ = chosen.(string)
	node.chosen = name + "." + node.chosen // A1.B1
	return node, nil
}

// A1 节点运行时从子节点中 选择 A1.B1 运行
func (node *TestNode6) Run(ctx context.Context) {
	fmt.Printf("Cluster %s running\n", node.Name())

	for _, subNode := range node.Base.SubNodes {
		if subNode.Name() == node.chosen {
			subNode.Run(ctx)
		}
	}

	fmt.Printf("Cluster %s stopped\n", node.Name())
}
```

utils 包提供了一些辅助方法可以简化读取 props ，如：

```go
helper := utils.ProxyProps(props)
node.wait = helper.SubGetString(name, "chosen")
```

#### State

state 用于存储 Node 运行过程中产生的数据，通过 state 可以实现 Node 之间的通信，state 是并发安全的。

相关定义如下：

```go
type Stateful interface {
	Node

	Bind(state State)
}

type State interface {
   Query(key string) (interface{}, bool)

   Update(key string, value interface{})

   Transform(key string, transform TransformStateFunc)
}

type TransformStateFunc func(from interface{}) interface{}
```

为 Node 实现 Bind 方法，引擎就会为 Node 绑定 state，这个 state 也会作为 ExecPlan 的执行结果输出，也就是上文提到的执行状态数据。

Query 方法用于查询某个键的值，Update 方法用于更新某个键的值，Transform 方法用于把旧值通过某种方式转换为新值，一般在需要部分更新某个键的值时使用，如在原数组基础上追加元素。

与 props 一样，utils 包提供了一些辅助方法可以简化读取 state

```go
value := utils.ProxyState(output.State).GetString("test_key")
```



#### running.Base

running 内置了一些基础接口实现，便于用户使用。比如 running.Base：

```go
type Base struct {
	NodeName string

	State State

	SubNodes []Node

	SubNodesMap map[string]Node
}

func (base *Base) SetName(name string) { // 辅助方法
	base.NodeName = name
}

func (base *Base) Name() string {
	return base.NodeName
}

func (base *Base) Inject(nodes []Node) {
	base.SubNodes = append(base.SubNodes, nodes...)

	if base.SubNodesMap == nil {
		base.SubNodesMap = make(map[string]Node)
	}

	for _, node := range nodes {
		base.SubNodesMap[node.Name()] = node
	}
}

func (base *Base) Bind(state State) {
	base.State = state

	for _, node := range base.SubNodes {
		if statefulNode, ok := node.(Stateful); ok {
			statefulNode.Bind(state)
		}
	}
}

func (base *Base) Run(ctx context.Context) { // 包含 Base 的 Node 必须实现 Run 方法
	panic("please implement run method")
}

func (base *Base) Reset() {
	base.ResetSubNodes()
}

func (base *Base) ResetSubNodes() { // 辅助方法
	for _, node := range base.SubNodes {
		node.Reset()
	}
}
```

在示例代码中也用到了它。在Node 中嵌入 running.Base 后，就不必再实现 SetName，Name，Inject，Bind方法，可以大量减少重复代码。

运行时直接通过 Base 获取 NodeName，State和 SubNodes。比如将TestNode1 修改为：

```go
type TestNode1 struct {
	running.Base
}

func (node *TestNode1) Run(ctx context.Context) {
	fmt.Printf("Single Node %s running\n", node.Base.NodeName)
	node.Base.State.Update("time", time.Now().Format("2006-01-02"))
	t, _ := node.Base.State.Query("time")
	fmt.Println(t)
	fmt.Printf("Single Node %s stopped\n", node.Base.NodeName)
}
```

则输出变为：

```
Single Node B1.A1 running
2022-05-08
Single Node B1.A1 stopped
```

其他内置实现见 common 包及对应 test 代码

## Roadmap

目前项目还在初期开发阶段，还有许多工作需要完成，这里先挖个坑

- [x] 增加测试代码
- [x] 支持更新 Plan
- [ ] 支持自定义 Engine，增加日志插件
- [ ] Plan 执行统计数据透出
- [ ] Worker 池优化

另外，项目的实现心得我会更新到博客上，👉[博客地址](https://symphony09.github.io/)
