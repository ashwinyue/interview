# 分布式系统面试原理详解

## 1. 分布式理论

### Q: 详细讲解 CAP 定理和 BASE 理论？

#### A: CAP 定理

```
┌─────────────────────────────────────────────────────────┐
│                    CAP 定理                              │
│                                                          │
│  一个分布式系统最多同时满足以下三项中的两项:              │
│                                                          │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐               │
│  │  C      │   │   A     │   │    P    │               │
│  │Consistency│Availability│Partition│               │
│  │  一致性  │   │  可用性  │   │ 容错性  │               │
│  └─────────┘   └─────────┘   └─────────┘               │
│                                                          │
│  P (分区容错) 是分布式系统的必然选择:                     │
│  - 网络分区必然发生                                      │
│  - 因此只能在 CA 和 CP 之间权衡                          │
│                                                          │
│  ┌─────────────────────┬──────────┬──────────┐          │
│  │      组合           │  放弃   │   选择   │          │
│  ├─────────────────────┼──────────┼──────────┤          │
│  │ CP (一致性 + 容错)  │   A     │ MongoDB │          │
│  │ AP (可用性 + 容错)  │   C     │ Cassandra│          │
│  │ CA (一致性 + 可用)  │   P     │   单机   │          │
│  └─────────────────────┴──────────┴──────────┘          │
└─────────────────────────────────────────────────────────┘
```

**详细解释：**

| 属性 | 说明 | 场景 |
|------|------|------|
| **C** Consistency | 所有节点在同一时间看到相同的数据 | 金融系统、库存扣减 |
| **A** Availability | 每个请求都能得到响应（成功/失败） | 社交网络、电商浏览 |
| **P** Partition Tolerance | 系统在网络分区时仍能继续运行 | 所有分布式系统 |

**一致性模型：**
```
强一致性 (Strong)
    │
    ├─── 线性一致性 (Linearizability)
    │     └─── 任何操作后立即可见
    │
    └─── 顺序一致性 (Sequential)
          └─── 操作按某种全局顺序排列

弱一致性 (Weak)
    │
    ├─── 最终一致性 (Eventual)
    │     └─── 系统保证最终达到一致
    │
    └─── 因果一致性 (Causal)
          └─── 有因果关系的操作保持顺序
```

---

#### BASE 理论

```
┌─────────────────────────────────────────────────────────┐
│                    BASE 理论                             │
│                                                          │
│  Basically Available (基本可用)                          │
│  - 系统允许出现局部故障，但不影响整体可用性                │
│  - 响应时间可以变慢，但不会一直等待                       │
│                                                          │
│  Soft State (软状态)                                     │
│  - 允许数据存在中间状态                                   │
│  - 这个中间状态不影响系统可用性                           │
│                                                          │
│  Eventually Consistent (最终一致性)                      │
│  - 经过一段时间后，所有副本最终达到一致                   │
│                                                          │
│  ┌──────────┐              ┌──────────┐                  │
│  │   ACID   │              │   BASE   │                  │
│  ├──────────┤              ├──────────┤                  │
│  │ 强一致性  │              │ 最终一致  │                  │
│  │  隔离性  │              │  基本可用  │                  │
│  │  持久性  │              │  软状态   │                  │
│  │ 传统数据库│              │ 分布式系统 │                  │
│  └──────────┘              └──────────┘                  │
└─────────────────────────────────────────────────────────┘
```

---

### Q: 什么是分布式一致性？如何实现？

#### A: 一致性级别

```
┌─────────────────────────────────────────────────────────┐
│                  一致性级别                              │
│                                                          │
│  强一致性                                                 │
│  │                                                         │
│  ├─── 读写一致性: Read after Write                       │
│  │     写后立即可读                                       │
│  │                                                         │
│  ├─── 会话一致性: Monotonic Reads                        │
│  │     同一会话内读操作单调                               │
│  │                                                         │
│  └─── 因果一致性: Causal Consistency                     │
│        因果相关的操作保持顺序                             │
│                                                          │
│  弱一致性                                                 │
│  │                                                         │
│  └─── 最终一致性: Eventual Consistency                   │
│        保证最终收敛                                       │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 分布式共识算法

### Q: Raft 协议的原理是什么？

#### A: Raft 核心机制

```
┌─────────────────────────────────────────────────────────┐
│                    Raft 状态机                            │
│                                                          │
│  每个节点有三种状态:                                      │
│                                                          │
│     Follower                                Leader       │
│        ○────────────◯ 超时选举                              │
│           │                                                │
│           │ 收到投票                                        │
│           ▼                                                │
│     Candidate                                          │
│        ○────────────◯ 获得多数票                           │
│           │                                                │
│           │                                                │
│           └──────────────▶ Leader                          │
│                              │                            │
│                              │ 收到更高Term                │
│                              └──▶ Follower                │
│                                                          │
│  任期 (Term): 递增的逻辑时钟                               │
│                                                          │
│  ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐                   │
│  │Term1│   │Term2│   │Term3│   │Term4│                   │
│  └─────┘   └─────┘   └─────┘   └─────┘                   │
│    Lead     Elect     Lead     Lead                      │
└─────────────────────────────────────────────────────────┘
```

---

#### Raft 两个子协议

**1. Leader 选举**
```
┌─────────────────────────────────────────────────────────┐
│                    选举流程                              │
│                                                          │
│  初始状态: 所有节点都是 Follower                          │
│                                                          │
│  Follower 选举超时 (150-300ms 随机)                       │
│        │                                                 │
│        ▼                                                 │
│  Candidate                                              │
│        │                                                 │
│        ├─── 给所有节点发送 RequestVote                   │
│        │                                                 │
│        ├─── 获得多数票 → 成为 Leader                      │
│        │                                                 │
│        └─── 未获得多数票 → 继续等待                       │
│                                                          │
│  投票规则:                                                │
│  1. Term 更大的投票                                      │
│  2. Term 相同, 日志更新的投票                            │
└─────────────────────────────────────────────────────────┘
```

**2. 日志复制**
```
┌─────────────────────────────────────────────────────────┐
│                    日志复制                              │
│                                                          │
│  Client                                                  │
│    │                                                     │
│    │ 写请求                                              │
│    ▼                                                     │
│  Leader                                                 │
│    │                                                     │
│    ├─── 追加本地日志                                     │
│    │                                                     │
│    └─── 并发 AppendEntries 给所有 Follower               │
│              │                                           │
│              ├─── Follower A (成功)                       │
│              ├─── Follower B (成功)                       │
│              └─── Follower C (失败) → 重试                │
│                                                          │
│  多数成功 → Commit → 返回 Client                         │
└─────────────────────────────────────────────────────────┘
```

---

#### Raft 日志格式

```go
type LogEntry struct {
    Index   int         // 日志索引
    Term    int         // 任期号
    Command []byte      // 操作命令
}

type Raft struct {
    currentTerm int        // 当前任期
    votedFor    int        // 投票给的候选人
    logs        []LogEntry // 日志条目
    commitIndex int        // 已提交的日志索引
    lastApplied int        // 已应用的日志索引

    // Leader 特有
    nextIndex  []int  // 每个节点的下一条日志索引
    matchIndex []int  // 每个节点已复制的日志索引
}
```

---

### Q: Paxos 和 Raft 的区别？

| 特性 | Paxos | Raft |
|------|-------|------|
| **易理解性** | 难懂 | 易懂（明确的设计目标） |
| **领导者** | 多 Leader 可能 | 单 Leader |
| **日志** | 无序 | 有序（简化） |
| **实现** | 复杂 | 相对简单 |
| **性能** | 理论上更优 | 实际应用更广 |

---

## 3. 分布式事务

### Q: 有哪些分布式事务解决方案？

#### A: 方案对比

```
┌─────────────────────────────────────────────────────────┐
│              分布式事务解决方案                          │
│                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │    2PC/3PC  │  │    TCC      │  │    SAGA     │     │
│  │  强一致性   │  │  最终一致   │  │  最终一致   │     │
│  │   阻塞     │  │   性能好   │  │   灵活     │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
│                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ 本地消息表  │  │ MQ 事务    │  │  最大努力    │     │
│  │  最终一致   │  │  最终一致   │  │  最终一致   │     │
│  │  可靠      │  │  异步      │  │  简单      │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
```

---

#### 2PC (两阶段提交)

```
┌─────────────────────────────────────────────────────────┐
│                   2PC 协议                               │
│                                                          │
│  阶段1: 准备阶段 (Prepare Phase)                         │
│  ┌────────────────────────────────────────┐              │
│  │  Coordinator ──▶ Participant A          │              │
│  │  "能执行事务吗?"  ──▶ Participant B    │              │
│  │                  ──▶ Participant C    │              │
│  │                                        │              │
│  │  ◀── "Yes" ─── Participant A          │              │
│  │  ◀── "Yes" ─── Participant B          │              │
│  │  ◀── "No"  ─── Participant C          │              │
│  │                                        │              │
│  │  任一 "No" → 回滚                      │              │
│  └────────────────────────────────────────┘              │
│                                                          │
│  阶段2: 提交阶段 (Commit Phase)                          │
│  ┌────────────────────────────────────────┐              │
│  │  Coordinator ──▶ "Commit!"            │              │
│  │                │                      │              │
│  │                ├──▶ Participant A     │              │
│  │                ├──▶ Participant B     │              │
│  │                └──▶ Participant C     │              │
│  │                                        │              │
│  │  ◀── "ACK" ─── Participant A          │              │
│  │  ◀── "ACK" ─── Participant B          │              │
│  │  ◀── "ACK" ─── Participant C          │              │
│  │                                        │              │
│  │  事务完成                              │              │
│  └────────────────────────────────────────┘              │
│                                                          │
│  问题:                                                   │
│  - 同步阻塞                                              │
│  - 单点故障                                              │
│  - 数据不一致（部分提交失败）                             │
└─────────────────────────────────────────────────────────┘
```

---

#### TCC (Try-Confirm-Cancel)

```
┌─────────────────────────────────────────────────────────┐
│                    TCC 模式                              │
│                                                          │
│  业务接口拆分为三个接口:                                 │
│                                                          │
│  Try:      尝试执行，预留资源                            │
│  Confirm: 确认执行，使用资源                             │
│  Cancel: 取消执行，释放资源                              │
│                                                          │
│  示例: 转账                                             │
│  ┌────────────────────────────────────────┐              │
│  │  Try:                                  │              │
│  │    A: 冻结 100 元 (余额 → 冻结)        │              │
│  │    B: 预留 100 元 (冻结 → 余额)        │              │
│  │                                        │              │
│  │  Confirm (全部成功):                   │              │
│  │    A: 扣除冻结余额                     │              │
│  │    B: 增加余额                         │              │
│  │                                        │              │
│  │  Cancel (任一失败):                    │              │
│  │    A: 解冻 (冻结 → 余额)               │              │
│  │    B: 释放预留 (余额 → 冻结)           │              │
│  └────────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────┘
```

**Go TCC 实现：**
```go
type TCC struct {
    tries    []TryFunc
    confirms []ConfirmFunc
    cancels  []CancelFunc
}

type TryFunc func() error
type ConfirmFunc func() error
type CancelFunc func() error

func (t *TCC) Execute() error {
    // 1. Try 阶段
    for _, try := range t.tries {
        if err := try(); err != nil {
            return err
        }
    }

    // 2. Confirm 阶段
    for i, confirm := range t.confirms {
        if err := confirm(); err != nil {
            // Confirm 失败，需要重试或人工介入
            go t.retryConfirm(i)
            return err
        }
    }
    return nil
}

func (t *TCC) Cancel() {
    // 逆序 Cancel
    for i := len(t.cancels) - 1; i >= 0; i-- {
        t.cancels[i]()
    }
}
```

---

#### SAGA (长活事务)

```
┌─────────────────────────────────────────────────────────┐
│                   SAGA 模式                             │
│                                                          │
│  将长事务拆分为多个本地事务，通过补偿机制保证一致性        │
│                                                          │
│  示例: 订单流程                                          │
│                                                          │
│  T1: 创建订单 ──▶ C1: 取消订单                          │
│      │                                                   │
│      ▼                                                   │
│  T2: 扣库存 ──▶ C2: 恢复库存                            │
│      │                                                   │
│      ▼                                                   │
│  T3: 支付 ──▶ C3: 退款                                  │
│      │                                                   │
│      ▼                                                   │
│  T4: 发货 ──▶ C4: 取消发货                              │
│                                                          │
│  正向: T1 → T2 → T3 → T4                                │
│  补偿: T3 失败 → C3 → C2 → C1 (逆序补偿)                 │
│                                                          │
│  两种模式:                                                │
│  - Choreography: 编舞模式 (事件驱动，无中心)             │
│  - Orchestration: 编排模式 (中心协调)                     │
└─────────────────────────────────────────────────────────┘
```

**Go SAGA 编排实现：**
```go
type Saga struct {
    steps []SagaStep
}

type SagaStep struct {
    Name    string
    Execute func(ctx context.Context) error
    Compensate func(ctx context.Context) error
}

func (s *Saga) Execute(ctx context.Context) error {
    executed := 0

    // 正向执行
    for i, step := range s.steps {
        if err := step.Execute(ctx); err != nil {
            // 失败，逆序补偿
            s.compensate(ctx, i-1)
            return err
        }
        executed = i + 1
    }
    return nil
}

func (s *Saga) compensate(ctx context.Context, from int) {
    for i := from; i >= 0; i-- {
        s.steps[i].Compensate(ctx)
    }
}

// 使用
func OrderSaga() *Saga {
    return &Saga{
        steps: []SagaStep{
            {
                Name: "create_order",
                Execute: func(ctx context.Context) error {
                    return CreateOrder(ctx)
                },
                Compensate: func(ctx context.Context) error {
                    return CancelOrder(ctx)
                },
            },
            {
                Name: "deduct_stock",
                Execute: func(ctx context.Context) error {
                    return DeductStock(ctx)
                },
                Compensate: func(ctx context.Context) error {
                    return RestoreStock(ctx)
                },
            },
            {
                Name: "payment",
                Execute: func(ctx context.Context) error {
                    return ProcessPayment(ctx)
                },
                Compensate: func(ctx context.Context) error {
                    return Refund(ctx)
                },
            },
        },
    }
}
```

---

## 4. 分布式锁

### Q: 如何实现一个分布式锁？

#### A: Redis 分布式锁

```
┌─────────────────────────────────────────────────────────┐
│              Redis 分布式锁原理                          │
│                                                          │
│  加锁:                                                   │
│  SET lock_key unique_value NX PX timeout                │
│  │    │         │            │  │                      │
│  │    │         │            │  过期时间               │
│  │    │         │            毫秒                      │
│  │    │         不存在才设置                             │
│  │    唯一标识 (用于解锁验证)                            │
│                                                          │
│  解锁:                                                   │
│  if redis.call("get", lock_key) == unique_value then     │
│      return redis.call("del", lock_key)                 │
│  end                                                     │
│  (Lua 脚本保证原子性)                                    │
└─────────────────────────────────────────────────────────┘
```

**Go 实现：**
```go
import "github.com/go-redis/redis/v8"

type RedisLock struct {
    client *redis.Client
    key    string
    value  string
    ttl    time.Duration
}

func NewRedisLock(client *redis.Client, key string, ttl time.Duration) *RedisLock {
    return &RedisLock{
        client: client,
        key:    key,
        value:  uuid.New().String(),
        ttl:    ttl,
    }
}

func (l *RedisLock) Lock(ctx context.Context) (bool, error) {
    return l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
}

func (l *RedisLock) Unlock(ctx context.Context) error {
    // Lua 脚本保证原子性
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        end
        return 0
    `
    return l.client.Eval(ctx, script, []string{l.key}, l.value).Err()
}

// 自动续期 (看门狗)
func (l *RedisLock) AutoRefresh(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            l.client.Expire(ctx, l.key, l.ttl)
        case <-ctx.Done():
            return
        }
    }
}
```

---

#### etcd 分布式锁

```go
import "go.etcd.io/etcd/client/v3/concurrency"

func etcdLock() error {
    cli, _ := clientv3.New(clientv3.Config{
        Endpoints: []string{"localhost:2379"},
    })
    defer cli.Close()

    session, _ := concurrency.NewSession(cli, concurrency.WithTTL(10))
    defer session.Close()

    mutex := concurrency.NewMutex(session, "/my-lock/")

    // 获取锁
    if err := mutex.Lock(context.TODO()); err != nil {
        return err
    }

    // 执行业务逻辑
    // ...

    // 释放锁
    return mutex.Unlock(context.TODO())
}
```

---

#### Redlock (Redis 集群锁)

```
┌─────────────────────────────────────────────────────────┐
│                   Redlock 算法                           │
│                                                          │
│  在 N 个独立的 Redis 节点上获取锁:                       │
│                                                          │
│  Client                                                  │
│    │                                                     │
│    ├───▶ Redis1 (SET NX PX) ──▶ Success                │
│    ├───▶ Redis2 (SET NX PX) ──▶ Success                │
│    ├───▶ Redis3 (SET NX PX) ──▶ Success                │
│    ├───▶ Redis4 (SET NX PX) ──▶ Success                │
│    └───▶ Redis5 (SET NX PX) ──▶ Success                │
│                                                          │
│  获取成功数 > N/2 → 锁获取成功                            │
│                                                          │
│  注意: 需要 N 个节点且独立部署（不同机器、不同机房）       │
└─────────────────────────────────────────────────────────┘
```

---

## 5. 分布式 ID

### Q: 如何生成全局唯一的分布式 ID？

#### A: 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **UUID** | 简单、无中心 | 无序、长度长 | 非主键 |
| **Snowflake** | 有序、高性能 | 依赖时钟 | 主键 |
| **数据库号段** | 可靠 | 性能低、依赖DB | 小规模 |
| **Redis incr** | 高性能 | 依赖Redis | 中规模 |
| **号段服务** | 高性能、可靠 | 复杂 | 大规模 |

---

#### Snowflake 算法

```
┌─────────────────────────────────────────────────────────┐
│                Snowflake ID 结构 (64bit)                 │
│                                                          │
│  ┌──┬────────┬────────┬────────────────┬─────────────┐  │
│  │0 │  timestamp │  worker  │   sequence    │           │
│  │  │  (41bit)  │ (10bit)  │    (12bit)    │           │
│  └──┴────────┴────────┴────────────────┴─────────────┘  │
│  1bit    41bit       10bit        12bit                  │
│                                                          │
│  - 符号位: 始终为0                                        │
│  - 时间戳: 毫秒级，可用69年                               │
│  - 机器ID: 5位数据中心 + 5位工作节点                      │
│  - 序列号: 毫秒内计数器                                  │
│                                                          │
│  QPS = 1000 * 4096 = 409万/秒 (单机)                    │
└─────────────────────────────────────────────────────────┘
```

**Go 实现：**
```go
type Snowflake struct {
    mu         sync.Mutex
    timestamp  int64
    workerID   int64
    sequence   int64
}

const (
    workerIDBits   = 10
    sequenceBits   = 12
    workerIDShift  = sequenceBits
    timestampShift = sequenceBits + workerIDBits
)

var epoch int64 = 1609459200000 // 2021-01-01 00:00:00

func NewSnowflake(workerID int64) *Snowflake {
    return &Snowflake{
        workerID: workerID,
    }
}

func (s *Snowflake) Generate() (int64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now().UnixMilli() - epoch

    if now == s.timestamp {
        s.sequence = (s.sequence + 1) & 0xFFF
        if s.sequence == 0 {
            // 序列号溢出，等待下一毫秒
            for now <= s.timestamp {
                time.Sleep(time.Microsecond)
                now = time.Now().UnixMilli() - epoch
            }
        }
    } else {
        s.sequence = 0
    }

    s.timestamp = now

    id := (now << timestampShift) |
          (s.workerID << workerIDShift) |
          s.sequence

    return id, nil
}
```

---

## 6. 服务治理

### Q: 服务注册发现、负载均衡、熔断降级如何实现？

#### A: 服务治理全景

```
┌─────────────────────────────────────────────────────────┐
│                   服务治理架构                            │
│                                                          │
│  ┌──────────────────────────────────────────────┐       │
│  │              Service Mesh (Istio/Linkerd)    │       │
│  │  ┌─────────┐   ┌─────────┐   ┌─────────┐    │       │
│  │  │ 流量管理 │   │  安全   │   │  可观测  │    │       │
│  │  └─────────┘   └─────────┘   └─────────┘    │       │
│  └──────────────────────────────────────────────┘       │
│                       │                                  │
│  ┌──────────────────────────────────────────────┐       │
│  │         应用层治理 (无 Mesh 场景)             │       │
│  │  ┌─────────┐   ┌─────────┐   ┌─────────┐    │       │
│  │  │Registry│   │  LB     │   │Circuit  │    │       │
│  │  │发现     │   │ 负载均衡 │   │ Breaker  │    │       │
│  │  └─────────┘   └─────────┘   └─────────┘    │       │
│  └──────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

---

#### 负载均衡策略

```go
type LoadBalancer interface {
    Next(services []string) (string, error)
}

// 随机
type RandomLB struct{}

func (r *RandomLB) Next(services []string) (string, error) {
    if len(services) == 0 {
        return "", ErrNoService
    }
    return services[rand.Intn(len(services))], nil
}

// 轮询
type RoundRobinLB struct {
    index uint32
}

func (r *RoundRobinLB) Next(services []string) (string, error) {
    if len(services) == 0 {
        return "", ErrNoService
    }
    i := atomic.AddUint32(&r.index, 1)
    return services[i%uint32(len(services))], nil
}

// 加权轮询
type WeightedRoundRobinLB struct {
    services []*WeightedService
}

type WeightedService struct {
    Name   string
    Weight int
    current int
}

func (w *WeightedRoundRobinLB) Next() (string, error) {
    total := 0
    maxIndex := -1
    maxCurrent := -1

    for i, s := range w.services {
        s.current += s.Weight
        total += s.Weight

        if s.current > maxCurrent {
            maxCurrent = s.current
            maxIndex = i
        }
    }

    if maxIndex >= 0 {
        w.services[maxIndex].current -= total
        return w.services[maxIndex].Name, nil
    }

    return "", ErrNoService
}

// 最少连接
type LeastConnLB struct {
    conns map[string]int
    mu     sync.RWMutex
}

func (l *LeastConnLB) Next(services []string) (string, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()

    minConns := int(^uint(0) >> 1)
    var selected string

    for _, s := range services {
        if l.conns[s] < minConns {
            minConns = l.conns[s]
            selected = s
        }
    }

    if selected != "" {
        l.conns[selected]++
        return selected, nil
    }

    return "", ErrNoService
}

func (l *LeastConnLB) Release(service string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.conns[service]--
}
```

---

## 7. 一致性 Hash

### Q: 什么是一致性 Hash？如何解决节点增减的问题？

#### A: 一致性 Hash 原理

```
┌─────────────────────────────────────────────────────────┐
│              一致性 Hash 环                              │
│                                                          │
│        0                                                 │
│     ┌─────┐                                             │
│   ┌─┴───────┐                                           │
│  │         │                                           │
│  ▼         │                                           │
│ Node A ────┼─────── Node C                              │
│            │       │                                    │
│            │       │                                    │
│         Node B  │                                       │
│              ──┴──                                       │
│                2^32-1                                   │
│                                                          │
│  数据映射:                                               │
│  hash(key) → 环上的点 → 顺时针最近的节点                  │
│                                                          │
│  节点增减:                                               │
│  - 新增 Node D: 只影响 D → 顺时针下一个节点之间的数据     │
│  - 删除 Node A: 只影响 A → 顺时针下一个节点之间的数据     │
└─────────────────────────────────────────────────────────┘
```

**虚拟节点解决数据倾斜：**
```
物理节点A         物理节点B
  │                │
  ▼                ▼
A0 ─ A1 ─ A2 ... B0 ─ B1 ─ B2 ...
  │     │        │     │
  └─────┴────────┴─────┘
         │
         ▼
    环上均匀分布
```

**Go 实现：**
```go
import "github.com/stathat/consistent"

type ConsistentHash struct {
    hash *consistent.Consistent
}

func NewConsistentHash(replicas int) *ConsistentHash {
    return &ConsistentHash{
        hash: consistent.New(),
    }
}

func (ch *ConsistentHash) Add(node string) {
    ch.hash.Add(node)
}

func (ch *ConsistentHash) Get(key string) (string, error) {
    return ch.hash.Get(key)
}

// 纯手写版本
type HashRing struct {
    nodes      []uint32
    nodeMap    map[uint32]string
    virtualNodes int
}

func NewHashRing(virtualNodes int) *HashRing {
    return &HashRing{
        nodeMap:     make(map[uint32]string),
        virtualNodes: virtualNodes,
    }
}

func (hr *HashRing) AddNode(node string) {
    for i := 0; i < hr.virtualNodes; i++ {
        virtualKey := fmt.Sprintf("%s#%d", node, i)
        hash := crc32.ChecksumIEEE([]byte(virtualKey))
        hr.nodes = append(hr.nodes, hash)
        hr.nodeMap[hash] = node
    }
    sort.Slice(hr.nodes, func(i, j int) bool {
        return hr.nodes[i] < hr.nodes[j]
    })
}

func (hr *HashRing) GetNode(key string) string {
    hash := crc32.ChecksumIEEE([]byte(key))

    // 二分查找
    idx := sort.Search(len(hr.nodes), func(i int) bool {
        return hr.nodes[i] >= hash
    })

    if idx == len(hr.nodes) {
        idx = 0 // 环形
    }

    return hr.nodeMap[hr.nodes[idx]]
}
```
