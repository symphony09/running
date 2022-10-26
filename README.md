# Running 一个简单的图节点执行框架

与其他图节点执行框架一样，running 也是将任务抽象为节点，将依赖关系抽象为边。

节点与边一起构成了图，也就是整个执行流程的抽象。

如何组织节点，按照预定的流程执行，就是 running 要解决的问题。

## 主要特点

### 注重并发执行场景

running 的设计注重许多任务需要并发执行的场景。所以它适合在线计算任务，而非离线计算任务。

例如 running 在执行过程中实时计算依赖，而非预先计算执行步骤，顺序执行。只要上游依赖解决，那么节点就会尽快运行。

此外 running 还着重考虑了冷启动问题，缓解瞬时高并发造成的性能问题。

### 支持定义复杂执行流程

在某些情况下，纯粹用图定义执行流程是比较困难的。

比如循环和分支选择流程，running 定义了簇 cluster 来解决这个问题。

cluster 是一类特殊的节点，它包含了若干个子节点，这些子节点的执行时机不再通过计算依赖关系确定，而是由 cluster 决定。

因为 cluster 也属于节点，所以它也可以嵌套在其他的 cluster 中，实现更加复杂的流程。

### 动态增强节点

将任务逻辑封装为节点，不可避免地会带来一定的心智负担。

running 定义了装饰器 wrapper 来动态增强节点，使用组合的理念降低负担。

一是增加节点的可复用性，避免需要频繁封装类似功能的节点。二是增加灵活性，新功能随用随加。

wrapper 也是一类特殊的节点，它包含要增强的目标节点，可以在目标节点执行前后执行增强逻辑。并且也可以嵌套。

### 方便地导出和载入图定义

running 支持对构建好的图进行序列化和反序列化，方便不同环境同步。

例如，工作流程在测试环境验证完成后，导出序列化图，线上环境热加载图定义就可以马上同步工作流程。

## 如何入手

### package common

common 包定义了一些通用的 cluster 、wrapper 和 node 实现。

结合对应 test 代码和方法注释就可以基本掌握如何使用 running。

十分建议在阅读完本文档后，阅读这些代码。有错误的地方也欢迎指出。

### running.Engine

running.Engine 提供了一组方法用于注册节点构建函数，管理计划（包括图定义和节点初始化参数）和 worker 池。

考虑到对 Engine 的定制需求较少，running 初始化了一个全局变量并暴露对应了方法。

因此可以直接调用 running.RegisterNodeBuilder 等函数，无需自行创建 Engine 实例来调用其方法。

有需要的话，如出于隔离目的，也可以使用 running.NewDefaultEngine 创建新的 Engine 实例。

### running.Plan

plan 包含了图定义和节点的初始化参数，图定义通过一系列操作选项确定。

将 plan 通过 RegisterPlan 注册到 engine 中后，就可以调用 ExecPlan 执行。

创建 plan 的函数签名如下：

```func NewPlan(props Props, prebuilt []Node, options ...Option) *Plan```

- props 即节点参数，running 提供了一个基于 map[string]interface{} 的实现 StandardProps。
  - key 基于约定设为 节点名.参数名，value 为参数值
- prebuilt 为预建节点，需要实现 Clone 方法，可以用于减少运行过程中节点构建消耗
- options 即操作选项，包括连接节点，合并节点为 cluster，包装节点等
  - 具体操作类型请参考 plan_options.go 及 test 代码

### running.Base & running.BaseWrapper

为了简化代码封装，running 提供 running.Base 和 running.BaseWrapper

将 running.Base 嵌入结构体后就自动实现了 Node 和 Cluster 接口的大部分方法，running.BaseWrapper 则实现 Wrapper 的接口方法，
只需再实现 Run 方法即可。 如果通用实现不满足要求，按需求重写对应方法即可。

### running.State

running.State 用于运行时节点间通信。只要实现 running.Stateful 接口, 节点在运行时就会被绑定 state。

通过对 state 增改查就能实现节点间的通信。

要注意的是，running 提供了 State 的一个实现 StandardState ，
StandardState 通过读写锁在一定程度上可以保证并发安全，但是前提是对于 Query 得到的对象不要做任何修改（尤其是引用类型）。
需要做修改时，使用 Update 方法或 Transform 方法。

在安全性或性能上有定制需求时，可以自行实现 State 接口，并设置 Engine 的 StateBuilder。

## 状态和路线图

目前项目任然处于初期阶段，不建议生产使用，同时欢迎提交 issue 来让 running 变得更好。

- [ ] 更加灵活的 props 操作