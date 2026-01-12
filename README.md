# Go 工程师面试准备指南

## 📋 学习计划概览

本指南按优先级和难度递进，建议按顺序学习。

---

## 第一阶段：Go 语言基础 (2-3周)

### 1.1 核心语法
- [ ] 变量声明与类型推断 (`:=`, `var`)
- [ ] 常量与 iota
- [ ] 基本类型：`int`, `int8/16/32/64`, `uint`, `float32/64`, `bool`, `string`
- [ ] 类型转换：Go 不支持隐式转换
- [ ] 控制结构：`if/else`, `switch`, `for` (Go 只有 for)
- [ ] 函数：多返回值、命名返回值、可变参数、defer
- [ ] error 处理模式

### 1.2 数据结构
- [ ] **Array vs Slice**
  - slice 底层数据结构 (ptr, len, cap)
  - slice 扩容机制
  - slice 作为函数参数的行为
- [ ] **Map**
  - map 底层实现 (hash table)
  - map 的线程安全问题
  - map 的遍历顺序
- [ ] **Struct**
  - 结构体定义与嵌入
  - 方法与接收者 (值 vs 指针)
  - 结构体比较
- [ ] **Interface**
  - 接口定义与实现 (duck typing)
  - 接口值的底层结构 (type, value)
  - 空接口 `interface{}`
  - 类型断言与 type switch

**面试高频题：**
```go
// 1. 切片扩容
s := make([]int, 0, 1)
s = append(s, 1)
s = append(s, 2)  // cap 会变成多少？

// 2. nil 切片 vs 空切片
var s1 []int           // nil
s2 := []int{}         // empty
len(s1) == 0  // ?
len(s2) == 0  // ?

// 3. map 线程安全问题
var m = make(map[int]int)
go func() { m[1] = 1 }()  // panic?

// 4. 接口值比较
var a interface{} = nil
var b interface{} = (*int)(nil)
a == b  // ?
```

---

## 第二阶段：并发编程 (3-4周)

### 2.1 Goroutine
- [ ] goroutine 调度模型 (GMP 模型)
- [ ] goroutine 泄漏场景
- [ ] goroutine 创建成本与数量控制

### 2.2 Channel
- [ ] channel 底层实现 (环形缓冲区)
- [ ] 有缓冲 vs 无缓冲 channel
- [ ] channel 操作：send、receive、close
- [ ] select 语句与超时处理
- [ ] channel 常见模式

### 2.3 Sync 包
- [ ] `sync.Mutex` vs `sync.RWMutex`
- [ ] `sync.WaitGroup`
- [ ] `sync.Once` (单例模式)
- [ ] `sync.Pool` (对象池)
- [ ] `sync/atomic` (原子操作)
- [ ] `sync.Cond` (条件变量)

### 2.4 Context
- [ ] context 的作用与传播
- [ ] context 类型：`Background`, `TODO`, `WithCancel`, `WithTimeout`, `WithValue`
- [ ] 超时控制与取消传播

**面试高频题：**
```go
// 1. channel 缓冲区
ch := make(chan int, 1)
ch <- 1
ch <- 2  // 会怎样？

// 2. select 随机性
select {
case ch1 <- 1:
case ch2 <- 2:
}  // 当两个 channel 都可用时？

// 3. goroutine 泄漏检测
func leak() {
    ch := make(chan int)
    go func() {
        <-ch  // 如果没有发送者会怎样？
    }()
}

// 4. Mutex 嵌套
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
// 在锁内再次 Lock 会怎样？
```

**常见并发模式：**
```go
// 1. Worker Pool 模式
// 2. Fan-Out / Fan-In 模式
// 3. Pipeline 模式
// 4. Publisher-Subscriber 模式
// 5. Rate Limiting (令牌桶/漏桶)
```

---

## 第三阶段：内存管理 (1-2周)

### 3.1 内存分配
- [ ] 栈 vs 堆分配
- [ ] 逃逸分析 (go build -gcflags="-m")
- [ ] TCMalloc 分配器

### 3.2 垃圾回收
- [ ] GC 算法演进：三色标记法
- [ ] 写屏障 (Write Barrier)
- [ ] GC 触发条件
- [ ] GOGC、GOMEMLIMIT 调优

### 3.3 内存优化
- [ ] 减少内存分配技巧
- [ ] sync.Pool 复用对象
- [ ] 避免常见内存泄漏

**面试高频题：**
```go
// 1. 逃逸分析
func foo() *int {
    i := 1  // i 会逃逸吗？
    return &i
}

// 2. map 持续增长导致的内存问题
var m = make(map[int][]byte)
for i := 0; i < 1000000; i++ {
    m[i] = make([]byte, 1024)
    delete(m, i)  // 内存会释放吗？
}
```

---

## 第四阶段：高级特性 (1-2周)

### 4.1 反射 (reflect)
- [ ] reflect 基本用法
- [ ] Type vs Value
- [ ] 反射的性能问题
- [ ] 常见应用场景

### 4.2 Unsafe
- [ ] unsafe.Pointer
- [ ] 指针运算
- [ ] 内存对齐
- [ ] 使用场景与风险

### 4.3 泛型 (Go 1.18+)
- [ ] 类型参数与约束
- [ ] 泛型类型与泛型函数
- [ ] 类型推断
- [ ] `comparable` 约束
- [ ] `any` vs `interface{}`

### 4.4 Go 新特性 (Go 1.21-1.23)
- [ ] `slices`, `maps` 包
- [ ] `log/slog` 结构化日志
- [ ] `cmp.Ordered` 比较
- [ ] range-over-func (实验性)

---

## 第五阶段：算法与数据结构 (持续练习)

### 5.1 必刷题目类型
- [ ] **数组/字符串**
  - 两数之和、三数之和
  - 最长无重复子串
  - 滑动窗口系列
- [ ] **链表**
  - 反转链表
  - 环形链表
  - 合并 K 个有序链表
- [ ] **二叉树**
  - 前中后序遍历
  - 层序遍历
  - 最近公共祖先
- [ ] **动态规划**
  - 斐波那契、爬楼梯
  - 背包问题
  - 最长公共子序列
- [ ] **排序与搜索**
  - 快排、归并排序
  - 二分搜索
  - 搜索旋转排序数组

### 5.2 刷题建议
```bash
# LeetCode Go 分类刷题
- Easy: 50 题
- Medium: 100 题
- Hard: 20-30 题

重点标签：
- Array (数组)
- Linked List (链表)
- Tree (树)
- Dynamic Programming (动态规划)
- Graph (图)
- Heap (堆)
```

---

## 第六阶段：系统设计 (2-3周)

### 6.1 设计原则
- [ ] SOLID 原则在 Go 中的实践
- [ ] Go 的设计哲学
  - "Communication is better than sharing memory"
  - "Don't communicate by sharing memory"

### 6.2 常见架构模式
- [ ] 分层架构 (Handler -> Service -> Repository)
- [ ] 整洁架构 (Clean Architecture)
- [ ] 六边形架构

### 6.3 微服务设计
- [ ] 服务拆分原则
- [ ] API 设计 (REST vs gRPC)
- [ ] 服务发现与负载均衡
- [ ] 熔断、降级、限流

### 6.4 存储设计
- [ ] 关系型数据库 (MySQL) 设计与优化
- [ ] NoSQL (Redis, MongoDB) 使用场景
- [ ] 数据一致性保证

---

## 第七阶段：分布式系统 (2周)

### 7.1 分布式理论
- [ ] CAP 定理
- [ ] BASE 理论
- [ ] 分布式事务 (2PC/3PC, TCC, SAGA, 本地消息表)
- [ ] 分布式锁 (Redis, etcd)
- [ ] 分布式 ID 生成

### 7.2 一致性与共识
- [ ] 一致性模型 (强一致、最终一致)
- [ ] Raft 协议
- [ ] Paxos 协议

### 7.3 服务治理
- [ ] 服务注册与发现 (Consul, etcd)
- [ ] 配置中心
- [ ] 链路追踪 (OpenTelemetry, Jaeger)
- [ ] 监控告警 (Prometheus, Grafana)

---

## 第八阶段：实战项目

### 项目建议

#### 项目一：并发安全缓存
```go
// 功能要求：
// - 支持 LRU 淘汰
// - 并发安全
// - 支持 TTL
// - 支持单机与分布式版本
```

#### 项目二：HTTP 服务框架
```go
// 功能要求：
// - 路由注册 (支持 /:id 参数)
// - 中间件机制
// - 超时控制
// - 优雅关闭
// - 限流与熔断
```

#### 项目三：分布式任务队列
```go
// 功能要求：
// - 基于 Redis
// - 任务重试机制
// - 延迟任务
// - 任务优先级
// - Worker Pool
```

---

## 面试高频知识点清单

### Go 语言特性
| 知识点 | 重要程度 |
|--------|----------|
| slice 底层与扩容 | ⭐⭐⭐⭐⭐ |
| map 线程安全 | ⭐⭐⭐⭐⭐ |
| interface 底层实现 | ⭐⭐⭐⭐⭐ |
| goroutine 调度 (GMP) | ⭐⭐⭐⭐⭐ |
| channel 底层实现 | ⭐⭐⭐⭐ |
| defer 执行顺序 | ⭐⭐⭐⭐ |
| GC 三色标记 | ⭐⭐⭐⭐ |
| context 使用场景 | ⭐⭐⭐⭐ |
| sync.Map vs map+sync.Mutex | ⭐⭐⭐ |
|逃逸分析 | ⭐⭐⭐ |

### 数据库与中间件
| 知识点 | 重要程度 |
|--------|----------|
| MySQL 索引原理 | ⭐⭐⭐⭐⭐ |
| 事务隔离级别 | ⭐⭐⭐⭐⭐ |
| Redis 数据结构 | ⭐⭐⭐⭐⭐ |
| Redis 持久化 (RDB/AOF) | ⭐⭐⭐⭐ |
| Kafka 消息可靠性 | ⭐⭐⭐⭐ |

---

## 学习资源推荐

### 书籍
1. 《Go 语言圣经》- 入门必读
2. 《Go 语言实战》- 实践导向
3. 《Go 语言底层原理剖析》- 深入理解
4. 《DDIA》(设计数据密集型应用) - 系统设计圣经

### 在线资源
- [Go 官方文档](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Blog](https://go.dev/blog/)

### 面试题库
- LeetCode (算法)
- 牛客网 (综合)
- 各公司面经 (牛客、知乎)

---

## 练习命令

```bash
# 初始化练习代码
go run main.go

# 运行测试
go test -v ./...

# 检查代码质量
go vet ./...

# 逃逸分析
go build -gcflags="-m" main.go

# CPU 性能分析
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof
go tool pprof mem.prof
```

---

## 每日学习计划建议

| 时间段 | 内容 |
|--------|------|
| 第1-3周 | Go 基础语法 + 数据结构 |
| 第4-6周 | 并发编程 + Channel 练习 |
| 第7-8周 | 内存管理 + 高级特性 |
| 第9-12周 | 算法刷题 (每天2-3题) |
| 第13-14周 | 系统设计 + 架构模式 |
| 第15-16周 | 分布式系统 + 中间件 |
| 第17-20周 | 实战项目 + 模拟面试 |

---

*祝你面试顺利！*
