# Go 并发编程面试原理详解

## 1. Goroutine 调度模型 (GMP)

### Q: 详细讲解 Go 的 GMP 调度模型？

#### A: GMP 模型核心组件

```
┌─────────────────────────────────────────────────────────┐
│                      M (Machine)                        │
│                    OS 线程                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │              P (Processor)                      │   │
│  │           逻辑处理器 / 调度上下文                │   │
│  │   ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐             │   │
│  │   │  G  │ │  G  │ │  G  │ │  G  │  本地队列    │   │
│  │   └─────┘ └─────┘ └─────┘ └─────┘             │   │
│  └─────────────────────────────────────────────────┘   │
│                      │                                 │
│              全局队列 (runq)                           │
└─────────────────────────────────────────────────────────┘
            ┌──────────┐
            │  网络 I/O │
            └──────────┘
```

**GMP 三要素：**

| 组件 | 全称 | 说明 | 数量 |
|------|------|------|------|
| **G** | Goroutine | 协程实体，包含栈、状态等 | 动态 |
| **M** | Machine | OS 线程，真正执行 G | ≤ 10,000 |
| **P** | Processor | 调度器，维护 G 本地队列 | = GOMAXPROCS |

---

#### 调度流程详解

**1. 调度循环 (schedule)**
```
M 从 P 获取任务:
1. P 的本地队列 → runqget()
2. 全局队列 → globrunqget() (每 61 次调度检查)
3. 网络轮询器 → netpoll (有 I/O 就绪)
4. 从其他 P 偷取 → stealWork()
```

**2. 工作窃取 (Work Stealing)**
```go
// 伪代码
func schedule() {
    // 1. 从本地队列获取
    if g := runqget(); g != nil {
        execute(g)
        return
    }

    // 2. 从全局队列获取 (每 61 次)
    if tick%61 == 0 {
        if g := globrunqget(); g != nil {
            execute(g)
            return
        }
    }

    // 3. 从其他 P 偷取
    for _, p := range allp {
        if g := runqsteal(p); g != nil {
            execute(g)
            return
        }
    }

    // 4. 网络轮询
    if g := netpoll(false); g != nil {
        execute(g)
        return
    }

    // 5. 没有工作，休眠 M
    stopm()
}
```

**3. 抢占式调度**

Go 1.14+ 基于信号的抢占式调度：

- **系统调用**: 执行超过 20 个 P 的 tick
- **函数调用**: 基于信号 `SIGURG` 抢占

```go
// 抢占检查位置
// 1. 函数入口
// 2. 循环回边 (loop back-edge)
// 3. GC 安全点
```

---

#### P 的数量设置

```go
// 默认: P 数量 = CPU 核心数
runtime.GOMAXPROCS(runtime.NumCPU())

// 设置为 1: 单核执行，用于测试
runtime.GOMAXPROCS(1)

// 设置为核心数: 充分利用多核
runtime.GOMAXPROCS(runtime.NumCPU())
```

**P 过少的危害：**
- M 会在空闲时阻塞，无法执行 G
- 导致并发度不足
- I/O 密集型任务尤甚

---

#### Goroutine 泄漏检测

```go
// 泄漏场景: 没有 receiver
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch  // 永远阻塞
        fmt.Println(val)
    }()
    // ch 没有 sender，goroutine 永远阻塞
}

// 正确做法
func noLeak() {
    ch := make(chan int, 1)  // buffered channel
    // 或者使用 context 取消
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return  // 正常退出
        }
    }()
    cancel()  // 取消 goroutine
}
```

---

## 2. Channel 底层原理

### Q: Channel 的底层实现？有缓冲和无缓冲的区别？

#### A: Channel 底层结构

```go
type hchan struct {
    qcount   uint           // 当前队列中元素个数
    dataqsiz uint           // 环形队列容量
    buf      unsafe.Pointer // 环形队列指针
    elemsize uint16         // 元素大小
    closed   uint32         // 是否关闭
    elemtype *_type         // 元素类型
    sendx    uint           // 发送索引
    recvx    uint           // 接收索引
    recvq    waitq          // 接收等待队列
    sendq    waitq          // 发送等待队列
    lock     mutex          // 互斥锁
}
```

**环形缓冲区设计：**
```
容量为 4 的 channel:

初始: sendx=0, recvx=0
┌───┬───┬───┬───┐
│   │   │   │   │
└───┴───┴───┴───┘
  ↑
 sendx/recvx

发送 1 个后: sendx=1, recvx=0
┌───┬───┬───┬───┐
│ A │   │   │   │
└───┴───┴───┴───┘
  ↑   ↑
recvx sendx

接收 1 个后: sendx=1, recvx=1
┌───┬───┬───┬───┐
│   │   │   │   │
└───┴───┴───┴───┘
    ↑
sendx=recvx
```

---

#### 发送/接收流程

**无缓冲 channel:**
```
发送流程:
1. 获取锁
2. 如果有接收者，直接发送，唤醒接收者
3. 否则，加入 sendq，阻塞当前 goroutine

接收流程:
1. 获取锁
2. 如果有发送者，直接接收，唤醒发送者
3. 否则，加入 recvq，阻塞当前 goroutine
```

**有缓冲 channel:**
```
发送流程:
1. 获取锁
2. 如果有接收者，直接发送
3. 否则如果队列未满，加入队列
4. 否则，加入 sendq，阻塞

接收流程:
1. 获取锁
2. 如果有发送者，直接接收
3. 否则如果队列不为空，从队列取出
4. 否则，加入 recvq，阻塞
```

---

#### select 实现原理

```go
select {
case ch1 <- v:
    // ...
case v := <-ch2:
    // ...
default:
    // ...
}
```

**编译后类似：**
```go
// 编译器为每个 case 生成 scase 结构
sel := make([]scase, 3)
sel[0].c = ch1
sel[0].elem = &v
sel[0].kind = caseSend
// ...

// 调用 runtime.selectgo
chosen, recvOK = selectgo(sel)
```

**selectgo 算法：**
1. **加锁**: 锁住所有涉及的 channel
2. **打乱顺序**: 随机化 case 顺序（保证公平）
3. **遍历**: 查找可执行的 case
4. **释放锁**: 解锁所有 channel
5. **执行**: 返回选中的 case

**多个 case 可用时随机选择：**
```go
ch1 := make(chan int, 1)
ch2 := make(chan int, 1)
ch1 <- 1
ch2 <- 2

// 每次执行结果不确定
select {
case <-ch1:
    fmt.Println("ch1")
case <-ch2:
    fmt.Println("ch2")
}
```

---

#### Channel 常见模式

**1. 超时控制:**
```go
select {
case result := <-ch:
    return result
case <-time.After(time.Second):
    return errors.New("timeout")
}
```

**2. 信号通知:**
```go
done := make(chan struct{})

// 发送信号
close(done)

// 等待信号
<-done
```

**3. 生产者-消费者:**
```go
func producer(ch chan<- int) {
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)
}

func consumer(ch <-chan int) {
    for v := range ch {
        fmt.Println(v)
    }
}
```

**4. 扇出/扇入:**
```go
// Fan-Out: 一个输入，多个处理
func fanOut(in <-chan int) []<-chan int {
    outs := make([]<-chan int, 3)
    for i := 0; i < 3; i++ {
        outs[i] = worker(in)
    }
    return outs
}

// Fan-In: 多个输入，一个输出
func fanIn(ins ...<-chan int) <-chan int {
    out := make(chan int)
    for _, in := range ins {
        go func(ch <-chan int) {
            for v := range ch {
                out <- v
            }
        }(in)
    }
    return out
}
```

---

## 3. Context 原理与使用

### Q: Context 的设计原理和使用场景？

#### A: Context 接口设计

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

**四种实现：**

| 类型 | 创建方式 | 用途 |
|------|----------|------|
| `emptyCtx` | `Background()`, `TODO()` | 根 context |
| `cancelCtx` | `WithCancel(parent)` | 可取消 |
| `timerCtx` | `WithTimeout()`, `WithDeadline()` | 超时取消 |
| `valueCtx` | `WithValue()` | 传递值 |

---

#### cancelCtx 原理

```go
type cancelCtx struct {
    Context

    mu       sync.Mutex
    done     chan struct{}
    children map[canceler]struct{}
    err      error
}

func (c *cancelCtx) Cancel(removeFromParent bool, err error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // 关闭 channel，通知所有等待者
    close(c.done)

    // 取消所有子 context
    for child := range c.children {
        child.cancel(false, err)
    }
    c.children = nil
}
```

**取消传播：**
```
root (Background)
 └── ctx1 (cancelCtx)
      ├── ctx2 (valueCtx)
      └── ctx3 (timeoutCtx)
           ├── ctx4
           └── ctx5

当 ctx3 超时时:
1. ctx3.Done() 关闭
2. ctx4, ctx5 的 Done() 也关闭
3. ctx1, ctx2 不受影响
```

---

#### Context 使用最佳实践

**1. 作为第一个参数传递：**
```go
func Do(ctx context.Context, arg string) error {
    // ...
}

// 调用
ctx := context.Background()
if err := Do(ctx, "hello"); err != nil {
    // ...
}
```

**2. 超时控制：**
```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := http.DefaultClient.Do(req)
```

**3. 传递请求范围数据：**
```go
type contextKey string

const userIDKey contextKey = "userID"

func handler(w http.ResponseWriter, r *http.Request) {
    ctx := context.WithValue(r.Context(), userIDKey, "12345")
    processUser(ctx)
}

func processUser(ctx context.Context) {
    userID := ctx.Value(userIDKey).(string)
    // ...
}
```

**4. 避免使用 context.Value 传递业务参数：**
```go
// 不推荐
ctx := context.WithValue(ctx, "name", "value")

// 推荐：使用函数参数
func process(ctx context.Context, name string) {}
```

---

## 4. Sync 包详解

### Q: sync.Map、sync.Pool、sync.Once 的使用场景？

#### A: sync.Map

**适用场景：**
1. 读多写少
2. key 集合稳定（不频繁增删）

```go
var m sync.Map

// 存
m.Store("key", "value")

// 取
if v, ok := m.Load("key"); ok {
    fmt.Println(v)
}

// 删
m.Delete("key")

// 原子操作
m.LoadOrStore("key", "default")
m.LoadAndDelete("key")
```

**底层原理：**
- `read`: atomic.Value，无锁读
- `dirty`: 普通 map，写操作
- `misses`: read 未命中计数，阈值 = len(dirty)

```
流程:
1. 先在 read 中查找（无锁）
2. 未找到且 amended=true，在 dirty 中查找（加锁）
3. misses 达到阈值，将 dirty 提升为 read
```

---

#### sync.Pool

**适用场景：**
1. 缓存临时对象，减少 GC 压力
2. 对象创建成本高（如 buffer）

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func getBuffer() []byte {
    return bufferPool.Get().([]byte)
}

func putBuffer(b []byte) {
    b = b[:0]  // 重置
    bufferPool.Put(b)
}
```

**特点：**
1. 每个 P 有自己的本地池
2. Get 时可能从其他 P 偷取
3. GC 时会清除 Pool 中的对象
4. 不适合存储连接等资源

---

#### sync.Once

**单例模式：**
```go
var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

**底层原理：**
```go
type Once struct {
    done uint32
    m    Mutex
}

func (o *Once) Do(f func()) {
    // 快速路径：原子检查
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```

**注意：** Do 传入的函数 panic 后，再次调用不会再执行（done 已设置）。

---

## 5. 常见并发模式

### Q: 实现常见的并发模式？

#### 1. Worker Pool

```go
func workerPool(tasks <-chan Task, numWorkers int) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                process(task)
            }
        }()
    }
    wg.Wait()
}
```

#### 2. Pipeline

```go
func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

// 使用
for n := range square(generator(1, 2, 3)) {
    fmt.Println(n)
}
```

#### 3. Or-Done

```go
func or(channels ...<-chan interface{}) <-chan interface{} {
    switch len(channels) {
    case 0:
        return nil
    case 1:
        return channels[0]
    }

    orDone := make(chan interface{})
    go func() {
        defer close(orDone)
        switch len(channels) {
        case 2:
            select {
            case <-channels[0]:
            case <-channels[1]:
            }
        default:
            select {
            case <-channels[0]:
            case <-channels[1]:
            case <-channels[2]:
            case <-or(channels[3:]...):
            }
        }
    }()
    return orDone
}
```

#### 4. Rate Limiter (令牌桶)

```go
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
    rl.tokens = min(rl.capacity, rl.tokens+int(elapsed*float64(rl.rate)))
    rl.lastTime = now

    if rl.tokens > 0 {
        rl.tokens--
        return true
    }
    return false
}
```
