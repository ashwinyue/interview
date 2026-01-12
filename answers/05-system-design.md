# 系统设计面试原理详解

## 1. 分层架构与设计原则

### Q: 请讲解常见的后端架构模式及其优缺点？

#### A: 三层/四层架构

```
┌─────────────────────────────────────────────────────────┐
│                     Client Layer                        │
│                   (Web/Mobile/API)                       │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   Presentation Layer                     │
│              (Handler/Controller/Router)                 │
│  - 参数校验                                              │
│  - 协议转换 (HTTP → Internal)                           │
│  - 统一错误处理                                          │
│  - 认证鉴权                                              │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                     Business Layer                       │
│                  (Service/Domain)                        │
│  - 业务逻辑                                              │
│  - 事务管理                                              │
│  - 业务规则验证                                          │
│  - 领域模型操作                                          │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   Persistence Layer                      │
│              (Repository/DAO/Mapper)                     │
│  - 数据访问                                              │
│  - 缓存操作                                              │
│  - SQL 构建                                              │
│  - 数据映射                                              │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              Database / Cache / External API             │
└─────────────────────────────────────────────────────────┘
```

**Go 实现示例：**
```go
// Handler 层
type UserHandler struct {
    userService *UserService
}

func (h *UserHandler) GetUser(c *gin.Context) {
    // 1. 参数解析
    id := c.Param("id")
    // 2. 参数校验
    if id == "" {
        c.JSON(400, gin.H{"error": "id required"})
        return
    }
    // 3. 调用 Service
    user, err := h.userService.GetUser(c.Request.Context(), id)
    if err != nil {
        // 4. 统一错误处理
        handleError(c, err)
        return
    }
    c.JSON(200, user)
}

// Service 层
type UserService struct {
    userRepo UserRepository
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // 业务逻辑
    if id == "admin" {
        return nil, ErrForbidden
    }
    // 调用 Repository
    return s.userRepo.FindByID(ctx, id)
}

// Repository 层
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
}
```

---

### Q: 什么是整洁架构 (Clean Architecture)？

#### A: 依赖倒置原则

```
外层 ─────────────────────────────────────────→ 内层
┌─────────────────────────────────────────────────────────┐
│                     Frameworks                          │
│              (Gin, gRPC, Redis Client)                  │
│              ────────┐                                   │
│                       │ 依赖                             │
│                       ▼                                   │
│              ┌───────────────────────┐                   │
│              │   Interface/Port      │  ◀── 定义接口      │
│              └───────────────────────┘                   │
│                       ▲                                  │
│                       │ 实现                             │
│              ┌────────┴────────┐                         │
│              │  Application   │                         │
│              │    (Use Cases)  │                         │
│              └─────────────────┘                         │
│                       ▲                                  │
│                       │ 依赖                             │
│              ┌────────┴────────┐                         │
│              │    Domain       │                         │
│              │  (Entities)     │  ◀── 核心业务逻辑       │
│              └─────────────────┘                         │
└─────────────────────────────────────────────────────────┘
```

**Go 实现示例：**
```go
// domain 层 - 核心业务实体（不依赖任何外部库）
type User struct {
    ID    string
    Name  string
    Email string
}

func (u *User) Validate() error {
    if u.Name == "" {
        return ErrInvalidName
    }
    return nil
}

// usecase 层 - 业务用例（定义接口）
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

type UserUseCase struct {
    repo UserRepository
}

func (uc *UserUseCase) CreateUser(ctx context.Context, name, email string) (*User, error) {
    user := &User{Name: name, Email: email}
    if err := user.Validate(); err != nil {
        return nil, err
    }
    return user, uc.repo.Save(ctx, user)
}

// infrastructure 层 - 具体实现（依赖 domain）
type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    // 具体实现
}
```

---

### Q: SOLID 原则在 Go 中如何实践？

#### A: SOLID 在 Go 中的体现

| 原则 | 说明 | Go 实践 |
|------|------|---------|
| **S**ingle Responsibility | 一个类只负责一件事 | 每个包/结构体职责单一 |
| **O**pen/Closed | 对扩展开放，对修改关闭 | 使用 interface |
| **L**iskov Substitution | 子类可替换父类 | 接口实现契约 |
| **I**nterface Segregation | 接口要小而专 | 小接口原则 |
| **D**ependency Inversion | 依赖抽象而非具体 | 依赖注入 |

```go
// S: 单一职责
// 错误: 一个结构体做太多事
type UserService struct{}
func (s *UserService) CreateUser() {}
func (s *UserService) SendEmail() {}
func (s *UserService) LogActivity() {}

// 正确: 拆分职责
type UserService struct{}
type EmailService struct{}
type ActivityLogService struct{}

// O: 开闭原则
type Logger interface {
    Log(msg string)
}

// 扩展: 添加新的日志实现，无需修改现有代码
type FileLogger struct{}
type KafkaLogger struct{}

// I: 接口隔离
// 错误: 胖接口
type UserInterface interface {
    Create()
    Update()
    Delete()
    Query()
    SendEmail()  // 不是所有实现都需要
}

// 正确: 小接口
type Creator interface {
    Create() error
}
type Sender interface {
    Send() error
}

// D: 依赖倒置
// 正确: 依赖接口
type OrderService struct {
    paymentProcessor PaymentProcessor  // 接口
    notifier         Notifier          // 接口
}
```

---

## 2. 微服务架构

### Q: 微服务拆分的原则是什么？

#### A: 拆分原则与维度

**拆分原则：**

```
┌─────────────────────────────────────────────────────────┐
│                   单体应用                               │
│              ┌──────────────────────┐                   │
│              │   Monolith App       │                   │
│              └──────────────────────┘                   │
└─────────────────────────────────────────────────────────┘
                          │
         ┌────────────────┼────────────────┐
         │                │                │
         ▼                ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   User Svc   │  │  Order Svc   │  │ Product Svc  │
└──────────────┘  └──────────────┘  └──────────────┘
```

**拆分维度：**

| 维度 | 说明 | 示例 |
|------|------|------|
| **业务领域** | 按业务边界 | 用户、订单、商品 |
| **功能模块** | 按功能分离 | 支付、通知、搜索 |
| **数据读写** | CQRS 模式 | 写服务、读服务 |
| **团队组织** | 康威定律 | 两个团队两个服务 |

**拆分时机：**
1. 单体应用开发效率下降
2. 部署风险增大
3. 团队规模扩大
4. 部分模块需要独立扩展

**拆分注意事项：**
```
┌─────────────────────────────────────────────────────────┐
│                    拆分前考虑                            │
│  1. 数据如何拆分？是否需要分布式事务？                   │
│  2. 服务间如何通信？同步还是异步？                       │
│  3. 如何保证数据一致性？                                 │
│  4. 如何处理分布式事务？                                 │
│  5. 运维复杂度是否可接受？                               │
│  6. 团队能力是否匹配？                                   │
└─────────────────────────────────────────────────────────┘
```

---

### Q: 服务间通信方式如何选择？

#### A: 同步 vs 异步

| 方式 | 协议 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|----------|
| **REST** | HTTP/JSON | 简单、通用 | 性能低 | 外部接口、简单调用 |
| **gRPC** | HTTP2/Protobuf | 高性能、类型安全 | 复杂 | 内部服务、高性能场景 |
| **消息队列** | 异步消息 | 解耦、削峰 | 复杂、延迟 | 异步任务、事件通知 |
| **共享存储** | Database | 简单 | 耦合 | 批量处理 |

**gRPC vs REST 性能对比：**
```
┌─────────────────────┬──────────┬──────────┐
│         metric       │   REST   │   gRPC   │
├─────────────────────┼──────────┼──────────┤
│ 序列化              │ JSON    │ Protobuf │
│ 序列化性能          │ 1x       │ 5x       │
│ 序列化后大小        │ 1x       │ 0.2x     │
│ 传输协议            │ HTTP/1.1 │ HTTP/2   │
│ 连接复用            │ 否       │ 是       │
│ 单向 QPS            │ ~10k     │ ~100k    │
└─────────────────────┼──────────┼──────────┘
```

**选择建议：**
```go
// 内部服务间通信 - 使用 gRPC
syntax = "proto3";
service OrderService {
    rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
    rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
}

// 对外接口 - 使用 REST
// 简单、通用、浏览器友好

// 异步事件 - 使用消息队列
// OrderCreated → 发送通知、更新库存、记录日志
```

---

### Q: 如何设计服务发现？

#### A: 服务发现方案

```
┌─────────────────────────────────────────────────────────┐
│                    服务发现架构                          │
│                                                          │
│   ┌─────────┐   ┌─────────┐   ┌─────────┐               │
│   │ Service│   │ Service│   │ Service│  服务注册         │
│   │   A    │   │   B    │   │   C    │                   │
│   └────┬────┘   └────┬────┘   └────┬────┘               │
│        │             │             │                     │
│        └─────────────┴─────────────┘                     │
│                      │                                   │
│                      ▼                                   │
│        ┌──────────────────────────┐                     │
│        │   Service Registry       │  注册中心           │
│        │  (Consul/etcd/Eureka)    │                     │
│        └─────────────┬────────────┘                     │
│                      │                                   │
│        ┌─────────────┴─────────────┐                     │
│        │                           │                     │
│   ┌────▼────┐                  ┌───▼────┐               │
│   │ Client  │                  │ Client │  服务发现      │
│   │   1     │                  │   2    │               │
│   └─────────┘                  └────────┘               │
└─────────────────────────────────────────────────────────┘
```

**方案对比：**

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **Consul** | 功能全面、UI好 | 复杂 | 中小规模 |
| **etcd** | 一致性强、K8s内置 | 功能少 | K8s环境 |
| **Nacos** | 阿里系、功能全 | Java味重 | Spring Cloud |
| **Eureka** | Netflix出品 | 维护少 | 旧项目 |
| **CoreDNS** | K8s原生 | 仅K8s | K8s环境 |

**Go 实现服务注册：**
```go
// Consul 注册示例
type ServiceRegistrar struct {
    client *consul.Client
}

func (r *ServiceRegistrar) Register(serviceID, name, address string, port int) error {
    registration := &consul.AgentServiceRegistration{
        ID:      serviceID,
        Name:    name,
        Address: address,
        Port:    port,
        Check: &consul.AgentServiceCheck{
            HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
            Interval:                       "10s",
            Timeout:                        "3s",
            DeregisterCriticalServiceAfter: "30s",
        },
    }
    return r.client.Agent().ServiceRegister(registration)
}

// 服务发现
func (r *ServiceRegistrar) Discover(serviceName string) (string, int, error) {
    services, _, err := r.client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return "", 0, err
    }
    // 负载均衡: 随机/轮询/最少连接
    svc := services[rand.Intn(len(services))]
    return svc.Service.Address, svc.Service.Port, nil
}
```

---

## 3. 缓存设计

### Q: 缓存穿透、击穿、雪崩如何解决？

#### A: 三大问题及解决方案

```
┌─────────────────────────────────────────────────────────┐
│                    缓存问题分类                          │
│                                                          │
│  1. 缓存穿透 (Penetration)                               │
│     ┌───┐    ┌─────┐    ┌──────┐                        │
│     │Req│───▶│Cache│    │      │                        │
│     └───┘    └─────┘    │ DB   │  查询不存在的数据       │
│                         │      │                        │
│     解决:                                        │
│     - 布隆过滤器                                        │
│     - 缓存空值                                          │
│                                                          │
│  2. 缓存击穿 (Breakdown)                                │
│     ┌───┐    ┌─────┐    ┌──────┐                        │
│     │Req│───▶│Cache│───▶│  DB  │  热点Key过期           │
│     └───┘    └─────┘    │      │                        │
│                         └──────┘                        │
│     解决:                                             │
│     - 互斥锁                                            │
│     - 永不过期                                          │
│     - 异步刷新                                          │
│                                                          │
│  3. 缓存雪崩 (Avalanche)                                │
│     ┌───┐    ┌─────┐    ┌──────┐                        │
│     │Req│───▶│Cache│    │  DB  │  大量Key同时过期        │
│     └───┘    └─────┘    │      │                        │
│                         └──────┘                        │
│     解决:                                             │
│     - 随机过期时间                                      │
│     - 多级缓存                                          │
│     - 限流降级                                          │
└─────────────────────────────────────────────────────────┘
```

**代码实现：**

```go
// 缓存穿透解决方案
type CacheWithBloom struct {
    cache    *redis.Client
    db       Database
    bloom    *BloomFilter
}

func (c *CacheWithBloom) Get(key string) (interface{}, error) {
    // 1. 布隆过滤器检查
    if !c.bloom.MightContain(key) {
        return nil, ErrNotFound
    }

    // 2. 查询缓存
    val, err := c.cache.Get(key).Result()
    if err == redis.Nil {
        // 3. 缓存未命中
        val, err = c.db.Query(key)
        if err != nil {
            // 4. 空值缓存（防止穿透）
            c.cache.Set(key, "", 5*time.Minute)
            return nil, err
        }
        c.cache.Set(key, val, 1*time.Hour)
        return val, nil
    }
    return val, nil
}

// 缓存击穿解决方案 - 单飞模式
type SingleFlight struct {
    group singleflight.Group
}

func (s *SingleFlight) Get(key string) (interface{}, error) {
    // 相同的 key 只有一次请求到达 DB
    val, err, _ := s.group.Do(key, func() (interface{}, error) {
        // 查询 DB
        return queryDB(key)
    })
    return val, err
}

// 缓存雪崩解决方案 - 随机过期
func SetWithRandomTTL(cache *redis.Client, key string, value interface{}, baseTTL time.Duration) {
    // 基础 TTL + 随机偏移 (±10%)
    offset := time.Duration(rand.Int63n(int64(baseTTL) / 10))
    ttl := baseTTL + offset
    cache.Set(key, value, ttl)
}
```

---

### Q: 如何设计多级缓存？

#### A: 两级缓存架构

```
┌─────────────────────────────────────────────────────────┐
│                    两级缓存架构                          │
│                                                          │
│  Request ──▶ L1 Cache (本地) ──▶ L2 Cache (Redis)       │
│              (BigCache)            │                     │
│                                    │ miss                │
│                                    ▼                     │
│                                 Database                │
│                                                          │
│  更新策略:                                                │
│  1. 更新 DB                                              │
│  2. 删除 L2                                              │
│  3. 广播删除 L1 (Pub/Sub)                                │
└─────────────────────────────────────────────────────────┘
```

**Go 实现：**
```go
type MultiLevelCache struct {
    l1 *Cache      // 本地缓存 (BigCache/sync.Map)
    l2 *redis.Client
    pubsub *redis.PubSub
}

func (m *MultiLevelCache) Get(key string) (interface{}, error) {
    // L1 查询
    if val, ok := m.l1.Get(key); ok {
        return val, nil
    }

    // L2 查询
    val, err := m.l2.Get(key).Result()
    if err == nil {
        m.l1.Set(key, val)
        return val, nil
    }

    // DB 查询
    val, err = m.queryDB(key)
    if err != nil {
        return nil, err
    }

    // 写入 L2, L1
    m.l2.Set(key, val, 1*time.Hour)
    m.l1.Set(key, val)
    return val, nil
}

func (m *MultiLevelCache) Set(key string, value interface{}) error {
    // 更新 DB
    if err := m.updateDB(key, value); err != nil {
        return err
    }

    // 删除缓存
    m.l1.Delete(key)
    m.l2.Del(key)

    // 通知其他节点删除 L1
    m.pubsub.Publish("cache:invalidate", key)
    return nil
}
```

---

## 4. 消息队列设计

### Q: 如何保证消息不丢失？

#### A: 三阶段保证

```
┌─────────────────────────────────────────────────────────┐
│                  消息流转三阶段                           │
│                                                          │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│  │  Producer│───▶│   Broker  │───▶│ Consumer │            │
│  └──────────┘    └──────────┘    └──────────┘           │
│       │               │               │                 │
│       ▼               ▼               ▼                 │
│   ┌───────┐       ┌───────┐       ┌───────┐             │
│   │ 发送确认 │       │ 持久化 │       │ 消费确认 │           │
│   │ Ack/  │       │  Sync  │       │  Ack  │             │
│   │ Callback│     │  Replica│      │ Manual │             │
│   └───────┘       └───────┘       └───────┘             │
└─────────────────────────────────────────────────────────┘
```

**Kafka 可靠配置：**
```go
// Producer 配置
producerConfig := sarama.NewConfig()
producerConfig.Producer.RequiredAcks = sarama.WaitForAll  // 等待所有副本确认
producerConfig.Producer.Retry.Max = 5                    // 重试5次
producerConfig.Producer.Return.Successes = true          // 启用成功回调

// 发送消息
msg := &sarama.ProducerMessage{
    Topic: "orders",
    Key:   sarama.StringEncoder(orderID),
    Value: sarama.ByteEncoder(data),
}
partition, offset, err := producer.SendMessage(msg)
// 返回分区和偏移量，表示发送成功

// Consumer 配置
consumerConfig := sarama.NewConfig()
consumerConfig.Consumer.Return.Errors = true
consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
consumerConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange

// 手动提交
consumer.MarkOffset(msg)          // 标记已处理
consumer.Commit()                 // 提交偏移量
```

---

### Q: 如何保证消息顺序？

#### A: 顺序消费方案

```
┌─────────────────────────────────────────────────────────┐
│                   消息顺序保证                            │
│                                                          │
│  问题: 多消费者打乱顺序                                   │
│                                                          │
│  Producer ──▶ [1,2,3,4,5,6,7,8,9]                       │
│       │                                                   │
│       ├───▶ Consumer A ──▶ [1,3,5,7,9]  乱序!           │
│       │                                                   │
│       └───▶ Consumer B ──▶ [2,4,6,8]    乱序!           │
│                                                          │
│  解决: 相同 Key 发送到同一分区                           │
│                                                          │
│  Producer (Key=UserID) ──▶ Partition0 ──▶ Consumer A   │
│                      └──▶ Partition1 ───▶ Consumer B   │
│                                                          │
│  同一用户消息保证顺序!                                   │
└─────────────────────────────────────────────────────────┘
```

**代码实现：**
```go
// 发送: 使用相同的 key 确保顺序
producer.SendMessage(&sarama.ProducerMessage{
    Topic: "orders",
    Key:   sarama.StringEncoder(userID),  // 同一用户 → 同一分区
    Value: orderData,
})

// 消费: 单个分区单线程
type OrderedConsumer struct {
    consumer sarama.ConsumerGroup
    handler  *OrderedHandler
}

type OrderedHandler struct {
    // 每个分区一个 channel
    partitions map[int32]chan *sarama.ConsumerMessage
    mu         sync.Mutex
}

func (h *OrderedHandler) Setup(session sarama.ConsumerGroupSession) error {
    h.mu.Lock()
    defer h.mu.Unlock()
    for _, partition := range session.Claims()["orders"] {
        h.partitions[partition] = make(chan *samed.ConsumerMessage, 100)
        go h.processPartition(partition)
    }
    return nil
}

func (h *OrderedHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for msg := range claim.Messages() {
        h.partitions[msg.Partition] <- msg
    }
    return nil
}

func (h *OrderedHandler) processPartition(partition int32) {
    for msg := range h.partitions[partition] {
        // 单线程顺序处理
        processMessage(msg)
    }
}
```

---

### Q: 如何处理消息积压？

#### A: 积压处理策略

```
┌─────────────────────────────────────────────────────────┐
│                    消息积压处理                          │
│                                                          │
│  检测: Consumer Lag 超过阈值                             │
│                                                          │
│  策略1: 横向扩展                                         │
│    增加消费者数量 (≤ 分区数)                              │
│                                                          │
│  策略2: 临时丢弃                                         │
│    只处理最新消息，历史数据离线处理                        │
│                                                          │
│  策略3: 重建 Topic                                      │
│    创建更多分区的新 Topic                                 │
│                                                          │
│  策略4: 优化消费逻辑                                     │
│    批量消费、异步处理、减少DB操作                         │
└─────────────────────────────────────────────────────────┘
```

**优化代码：**
```go
// 批量消费
func (c *Consumer) ConsumeBatch(batchSize int) {
    batch := make([]*Message, 0, batchSize)
    timeout := time.NewTimer(100 * time.Millisecond)

    for len(batch) < batchSize {
        select {
        case msg := <-c.messages:
            batch = append(batch, msg)
        case <-timeout.C:
            goto process
        }
    }

process:
    // 批量写入 DB
    c.db.BatchInsert(batch)
}

// 异步处理 + 结果队列
func (c *Consumer) AsyncProcess() {
    results := make(chan error, 100)

    for msg := range c.messages {
        msg := msg
        go func() {
            results <- processMessage(msg)
        }()
    }

    // 等待结果
    for i := 0; i < cap(results); i++ {
        if err := <-results; err != nil {
            // 错误处理
        }
    }
}
```

---

## 5. 数据库设计

### Q: 如何设计分库分表？

#### A: 分片策略

```
┌─────────────────────────────────────────────────────────┐
│                   分库分表策略                            │
│                                                          │
│  1. 水平分片 (Sharding)                                  │
│     ┌──────────────────────────────────────┐            │
│     │          原表 (1亿行)                │            │
│     └──────────────────────────────────────┘            │
│              │ 分片规则                                   │
│     ┌────────┼────────┬────────┬────────┐               │
│     ▼        ▼        ▼        ▼        ▼               │
│  ┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐             │
│  │Shard0││Shard1││Shard2││Shard3││Shard4│             │
│  │2000w ││2000w ││2000w ││2000w ││2000w │             │
│  └──────┘└──────┘└──────┘└──────┘└──────┘             │
│                                                          │
│  分片规则:                                                │
│  - Hash: hash(user_id) % N                              │
│  - Range: user_id 在 [0, 1000万) → Shard0               │
│  - 地理位置: 按城市/区域                                 │
│                                                          │
│  2. 垂直分片                                           │
│  ┌─────────────────────┐    ┌─────────────────────┐     │
│  │   用户基础信息       │    │   用户扩展信息       │     │
│  │ (id, name, email)  │    │ (profile, settings) │     │
│  └─────────────────────┘    └─────────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

**Go 分片实现：**
```go
type ShardingDB struct {
    shards []*sql.DB
    count  int
}

func (s *ShardingDB) GetShard(key string) *sql.DB {
    // Hash 分片
    hash := crc32.ChecksumIEEE([]byte(key))
    index := hash % uint32(s.count)
    return s.shards[index]
}

func (s *ShardingDB) QueryUser(userID string) (*User, error) {
    db := s.GetShard(userID)
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    return &user, err
}

// 分片表名生成
func (s *ShardingDB) GetTableName(userID string) string {
    hash := crc32.ChecksumIEEE([]byte(key))
    shard := hash % 100
    return fmt.Sprintf("users_%02d", shard)
}
```

---

### Q: 读写分离如何实现？

#### A: 读写分离架构

```
┌─────────────────────────────────────────────────────────┐
│                    读写分离架构                          │
│                                                          │
│             Application                                 │
│        ┌─────────────────────┐                          │
│        │   ReadWrite Split   │                          │
│        │      Middleware     │                          │
│        └───────┬─────┬───────┘                          │
│                │     │                                  │
│         Write  │     │  Read                            │
│                ▼     ▼                                  │
│         ┌──────────┐  ┌──────────┐                     │
│         │ Master   │  │  Slave   │                     │
│         │ (Primary)│  │(Replica) │                     │
│         └────┬─────┘  └────┬─────┘                     │
│              │             │                            │
│              └─── 复制 ────┘                            │
│                                                          │
│  中间件: ProxySQL / MaxScale / 自研                     │
└─────────────────────────────────────────────────────────┘
```

**代码实现：**
```go
type ReadWriteSplitDB struct {
    master *sql.DB
    slaves []*sql.DB
    index  uint64
}

func (rw *ReadWriteSplitDB) Master() *sql.DB {
    return rw.master
}

func (rw *ReadWriteSplitDB) Slave() *sql.DB {
    // 轮询负载均衡
    i := atomic.AddUint64(&rw.index, 1)
    return rw.slaves[i%uint64(len(rw.slaves))]
}

// 事务强制走 Master
func (rw *ReadWriteSplitDB) Begin() (*sql.Tx, error) {
    return rw.master.Begin()
}

// 读操作走 Slave
func (rw *ReadWriteSplitDB) QueryUser(id string) (*User, error) {
    db := rw.Slave()
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    return &user, err
}

// 写操作走 Master
func (rw *ReadWriteSplitDB) SaveUser(user *User) error {
    _, err := rw.master.Exec("INSERT INTO users ...")
    return err
}
```

---

## 6. 限流、熔断、降级

### Q: 如何实现限流、熔断、降级？

#### A: 三种保护机制

```
┌─────────────────────────────────────────────────────────┐
│                  服务保护机制                            │
│                                                          │
│  Request ──▶ RateLimiter ──▶ CircuitBreaker ──▶ Service  │
│              (限流)              (熔断)                  │
│               │                   │                      │
│               │                   │                      │
│               ▼                   ▼                      │
│            直接返回              Fallback                │
│            "Too Many"            (降级)                  │
└─────────────────────────────────────────────────────────┘
```

**代码实现：**
```go
// 限流器 - 令牌桶
type RateLimiter struct {
    rate     int
    capacity int
    tokens   int
    lastTime time.Time
    mu       sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(rl.lastTime).Seconds()
    rl.tokens = min(rl.capacity, int(elapsed*float64(rl.rate)))

    if rl.tokens > 0 {
        rl.tokens--
        return true
    }
    return false
}

// 熔断器
type CircuitBreaker struct {
    maxFailures int
    resetTime   time.Duration
    mu          sync.Mutex
    failures    int
    lastFail    time.Time
    state       State
}

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()

    // 检查是否需要重置
    if cb.state == StateOpen && time.Since(cb.lastFail) > cb.resetTime {
        cb.state = StateHalfOpen
        cb.failures = 0
    }

    // 熔断开启，直接返回
    if cb.state == StateOpen {
        cb.mu.Unlock()
        return ErrCircuitOpen
    }

    cb.mu.Unlock()

    // 执行函数
    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFail = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
        return err
    }

    // 半开状态成功则关闭
    if cb.state == StateHalfOpen {
        cb.state = StateClosed
        cb.failures = 0
    }
    return nil
}

// 降级处理
func GetUserInfoWithFallback(userID string) (*User, error) {
    // 尝试从缓存获取
    if user, err := cache.Get(userID); err == nil {
        return user, nil
    }

    // 调用服务
    if user, err := userService.Get(userID); err == nil {
        return user, nil
    }

    // 降级: 返回默认值或旧数据
    return &User{ID: userID, Name: "Unknown"}, nil
}
```

---

## 7. 监控与链路追踪

### Q: 如何设计可观测性？

#### A: 三大支柱

```
┌─────────────────────────────────────────────────────────┐
│                   可观测性三大支柱                        │
│                                                          │
│  1. Metrics (指标) - 聚合数据                             │
│     ┌────────────┐  ┌────────────┐  ┌────────────┐     │
│     │ RED Method │  │ Four Golden │  │ USE Method │     │
│     │ Rate       │  │   Signals   │  │            │     │
│     │ Errors     │  │   Latency   │  │ Utilization│     │
│     │ Duration   │  │   Traffic   │  │ Saturation │     │
│     └────────────┘  │   Errors    │  │ Errors     │     │
│                     └────────────┘  └────────────┘     │
│                                                          │
│  2. Logs (日志) - 离散事件                               │
│     结构化日志 (JSON) + 日志聚合 (ELK/Loki)              │
│                                                          │
│  3. Traces (链路) - 请求路径                              │
│     Distributed Tracing (OpenTelemetry + Jaeger)         │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Go Prometheus 指标：**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
    prometheus.MustRegister(requestDuration)
}

// Middleware
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())

        requestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

**OpenTelemetry 链路追踪：**
```go
import "go.opentelemetry.io/otel"

func tracerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tracer := otel.Tracer("gin-server")
        ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
        defer span.End()

        // 添加属性
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
        )

        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```
