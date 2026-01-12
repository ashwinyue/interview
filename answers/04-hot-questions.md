# Go 高频面试题速查手册

## 100+ 道精选面试题，快速复习

---

## 一、语言基础

### 1. new vs make
| | new | make |
|---|-----|------|
| 返回 | *T | T |
| 适用 | 所有类型 | map/slice/channel |
| 初始化 | 零值 | 指定值 |

```go
p1 := new(int)   // *int, 值为 0
p2 := &int{}     // *int, 值为 0
m1 := make(map[int]int)  // map[int]int
m2 := map[int]int{}      // map[int]int
```

---

### 2. cap 的含义
```go
var a [10]int           // cap(a) = 10
s := a[0:5]             // cap(s) = 10 (从开始位置到底层数组末尾)
s = a[3:7]              // cap(s) = 7

c := make(chan int, 5)  // cap(c) = 5 (缓冲大小)
m := make(map[int]int)  // cap(m) = ? (未定义)
```

---

### 3. rune 类型
```go
// rune 是 int32 的别名，表示 Unicode 码点
s := "你好"
fmt.Println(len(s))      // 6 (字节数)
fmt.Println(len([]rune(s))) // 2 (字符数)

// 遍历中文
for _, r := range s {  // r 是 rune
    fmt.Printf("%c", r)
}
```

---

### 4. 零值 vs nil
```go
// 有 nil 的类型: pointer, slice, map, channel, interface, function
// 无 nil 的类型: int, float, bool, string, array, struct

var s []int         // s == nil
var m map[int]int   // m == nil
var c chan int      // c == nil
var f func()        // f == nil
var i interface{}   // i == nil

var arr [10]int     // 无 nil
var num int         // 无 nil
```

---

### 5. 字符串不可变
```go
s := "hello"
// s[0] = 'H'  // 编译错误
s = "Hello"       // OK，创建新字符串

// 修改需要转换
b := []byte(s)
b[0] = 'H'
s = string(b)
```

---

### 6. 数组 vs 切片
```go
// 数组: 值类型，大小固定，比较用 ==
a1 := [3]int{1, 2, 3}
a2 := [3]int{1, 2, 3}
fmt.Println(a1 == a2)  // true

// 切片: 引用类型，大小可变，不能用 == 比较
s1 := []int{1, 2, 3}
s2 := []int{1, 2, 3}
// s1 == s2  // 编译错误
```

---

### 7. map 遍历顺序
```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v)  // 顺序随机
}
// Go 1.0-1.16: 顺序随机（防止依赖顺序）
// Go 1.17+: 单次运行内顺序固定，每次运行不同
```

---

### 8. map key 的类型
```go
// 可比较类型都能做 key: int, float, string, bool, pointer, array, channel, interface
// 不可比较: slice, map, function

// 正确
map[string]int
map[[3]int]string
map[interface{}]int

// 错误
map[[]int]string
map[map[int]int]bool
```

---

### 9. 结构体比较
```go
// 可比较: 所有字段都可比较
type A struct {
    X int
    Y string
}
a1, a2 := A{1, "x"}, A{1, "x"}
fmt.Println(a1 == a2)  // true

// 不可比较: 包含不可比较字段
type B struct {
    m map[int]int
}
b1, b2 := B{}, B{}
// b1 == b2  // 编译错误
```

---

### 10. 方法集
```go
type T struct{}
type S struct{}

func (t T) M1() {}  // 值接收者
func (t *T) M2() {} // 指针接收者

// T 的方法集: {M1, M2}
// *T 的方法集: {M1, M2}

var t T
t.M1()  // OK
t.M2()  // OK (自动取地址)

var p *T
// p.M1()  // 编译错误! nil 指针
p.M2()  // OK

// 接口实现规则:
// - T 实现的接口，*T 也实现
// - *T 实现的接口，T 不一定实现
```

---

## 二、并发编程

### 11. goroutine vs 线程
| 特性 | goroutine | 线程 |
|------|-----------|------|
| 创建成本 | ~2KB | ~2MB |
| 调度 | 用户态 (Go) | 内核态 (OS) |
| 切换成本 | 低 | 高 |
| 数量 | 百万级 | 千级 |

---

### 12. channel 方向
```go
var ch1 chan int           // 双向
var ch2 chan<- int         // 只发送
var ch3 <-chan int         // 只接收

// 类型转换
ch1 = make(chan int)
ch2 = ch1       // OK, 双向→单向
// ch1 = ch2     // 编译错误, 单向→双向
```

---

### 13. range channel
```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch)

for v := range ch {
    fmt.Println(v)  // 1, 2
}
// channel 关闭后，range 自动退出
// 如果不关闭，range 会永远阻塞
```

---

### 14. select 随机性
```go
ch1, ch2 := make(chan int, 1), make(chan int, 1)
ch1 <- 1
ch2 <- 2

for i := 0; i < 100; i++ {
    select {
    case <-ch1:
        // 随机执行
    case <-ch2:
        // 随机执行
    }
}
// 多个 case 就绪时，随机选择一个
```

---

### 15. 空 select
```go
select {}  // 永远阻塞

// 用于:
// 1. 阻止 main 返回
func main() {
    // ...
    select {}
}

// 2. 永远等待信号
select {}
```

---

### 16. 缓冲 channel 用法
```go
// 1. 信号量
sem := make(chan struct{}, 10)
for i := 0; i < 100; i++ {
    sem <- struct{}{}
    go func() {
        // ... 工作
        <-sem
    }()
}

// 2. 异步日志
log := make(chan LogEntry, 1000)
go func() {
    for entry := range log {
        writeLog(entry)
    }
}()
```

---

### 17. sync.Mutex vs sync.RWMutex
```go
// Mutex: 一次只有一个读写
var mu sync.Mutex
mu.Lock()
// 读写
mu.Unlock()

// RWMutex: 多读单写
var rwmu sync.RWMutex
rwmu.RLock()  // 读锁，多个读者可以同时持有
// 读
rwmu.RUnlock()

rwmu.Lock()   // 写锁，排他
// 写
rwmu.Unlock()
```

---

### 18. WaitGroup 常见错误
```go
// 错误: 复制 WaitGroup
wg := sync.WaitGroup{}
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(wg sync.WaitGroup) {  // 复制了!
        defer wg.Done()
        // ...
    }(wg)
}
// panic: WaitGroup 在错误状态下使用

// 正确: 传指针
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(wg *sync.WaitGroup) {
        defer wg.Done()
        // ...
    }(&wg)
}
```

---

### 19. context 取消传播
```go
ctx, cancel := context.WithCancel(context.Background())
ctx1, cancel1 := context.WithCancel(ctx)
ctx2, cancel2 := context.WithCancel(ctx)

cancel()  // 取消 ctx

// ctx1, ctx2 也被取消
<-ctx1.Done()  // 不阻塞
<-ctx2.Done()  // 不阻塞

// 但 cancel1 不影响 ctx2
cancel1()     // 只取消 ctx1
<-ctx2.Done()  // 不受影响
```

---

### 20. Context 超时
```go
// WithTimeout: 相对时间
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

// WithDeadline: 绝对时间
ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
defer cancel()

// 超时后 ctx.Err() == context.DeadlineExceeded
```

---

## 三、内存与GC

### 21. 堆 vs 栈
```go
// 栈分配: 函数内的局部变量
func f() {
    x := 42  // 通常在栈上
}

// 堆分配: 逃逸的变量
func f() *int {
    x := 42
    return &x  // x 逃逸到堆
}
```

---

### 22. GOGC 含义
```go
// GOGC 控制触发 GC 的堆增长百分比
// GOGC=100 (默认): 堆增长 100% 触发
// GOGC=200: 堆增长 200% 触发
// GOGC=off: 禁用 GC

// 程序中设置
debug.SetGCPercent(200)
```

---

### 23. 内存泄漏排查
```bash
# 1. 查看 runtime.MemStats
# 2. pprof
go tool pprof http://localhost:8080/debug/pprof/heap

# 3. goroutine 泄漏
go tool pprof http://localhost:8080/debug/pprof/goroutine

# 4. 实时监控
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -http=:8081 heap.prof
```

---

### 24. 大对象分配
```go
// > 32KB 直接从 mheap 分配
// < 32KB 从 mcache/mcentral 分配

large := make([]byte, 1024*1024)  // 1MB, 直接堆分配
small := make([]byte, 1024)       // 1KB, 可能栈分配
```

---

### 25. 减少分配技巧
```go
// 1. 预分配
s := make([]int, 0, 1000)

// 2. 复用 sync.Pool
var pool = sync.Pool{New: func() interface{} {
    return make([]byte, 1024)
}}

// 3. 避免不必要的指针
type Point struct {
    X, Y int  // 好
}
type Point struct {
    X, Y *int // 差
}

// 4. 字符串拼接用 strings.Builder
var b strings.Builder
b.Grow(100)
b.WriteString("hello")
```

---

## 四、接口与反射

### 26. 接口比较
```go
// 只有两个都是 nil 时才相等
var i1 interface{}
var i2 interface{}
fmt.Println(i1 == i2)  // true

var p *int = nil
i1 = p
fmt.Println(i1 == nil)  // false!
// i1 的 type 是 *int，value 是 nil
```

---

### 27. 类型断言
```go
var i interface{} = "hello"

// 安全方式
if s, ok := i.(string); ok {
    fmt.Println(s)
}

// 不安全方式，panic
s := i.(string)

// type switch
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown")
}
```

---

### 28. 空接口 vs any
```go
// 完全相同
var i1 interface{}
var i2 any

// any 是 Go 1.18 引入的别名
// 更简洁，推荐使用
```

---

### 29. 反射性能
```go
// 反射比直接调用慢约 100 倍
// 尽量避免热路径使用反射

// 优化: 缓存 reflect.Value
var cachedType = reflect.TypeOf(MyType{})

func process(v interface{}) {
    rv := reflect.ValueOf(v)
    // 使用 rv
}
```

---

### 30. json.Unmarshal 原理
```go
// 使用反射解析字段
// 支持标签: json:"field,name,omitempty"

type User struct {
    Name string `json:"name"`
    Age  int    `json:"age,omitempty"`
    // omitempty: 零值不序列化
}

// Unmarshal 流程:
// 1. 解析 JSON
// 2. 反射获取字段类型
// 3. 根据类型转换值
// 4. 设置字段值
```

---

## 五、标准库

### 31. time.Tick 用法
```go
// 正确: 用 range
for t := range time.Tick(time.Second) {
    fmt.Println(t)
}

// 错误: Tick 会永远泄漏 ticker
ticker := time.NewTicker(time.Second)
defer ticker.Stop()
for t := range ticker.C {
    fmt.Println(t)
}
```

---

### 32. http.Client 复用
```go
// 正确: 复用 Client
var client = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
}

// 错误: 每次创建新 Client
resp, err := http.Get(url)  // 使用默认 Client
```

---

### 33. io.Copy 最佳实践
```go
// io.Copy 自动使用缓冲
io.Copy(dst, src)

// 有大小限制
io.CopyN(dst, src, 1024)

// 读写都缓冲
bufio.NewReader(src)
bufio.NewWriter(dst)
```

---

### 34. log.Fatal
```go
// log.Fatal 会调用 os.Exit(1)
// defer 不会执行!

func f() {
    defer fmt.Println("cleanup")
    log.Fatal("error")  // 直接退出，defer 不执行
}

// 正确: 返回 error
func f() error {
    return errors.New("error")
}
```

---

### 35. errors.New vs fmt.Errorf
```go
// 创建新错误
err1 := errors.New("something wrong")

// 格式化错误
err2 := fmt.Errorf("context: %v", err1)

// Go 1.13+ 错误包装
err3 := fmt.Errorf("wrap: %w", err1)
// errors.Is(err3, err1)  // true
// errors.Unwrap(err3)    // err1
```

---

## 六、高级特性

### 36. 泛型约束
```go
// 有序类型约束
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
        ~float32 | ~float64 | ~string
}

func Min[T Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// ~ 表示底层类型匹配
type MyInt int
var mi MyInt = 10
Min(mi, 20)  // OK
```

---

### 37. comparable 约束
```go
// comparable: 可用 == 比较
func Index[T comparable](slice []T, x T) int {
    for i, v := range slice {
        if v == x {  // comparable 保证可以用 ==
            return i
        }
    }
    return -1
}

// 不能用: slice, map, func
Index([]int{1, 2, 3}, 2)      // OK
Index([]func(){}, func(){})  // 编译错误
```

---

### 38. any 类型
```go
// any 是 interface{} 的别名
func f(v any) {}

// 两者完全相同
var i1 interface{}
var i2 any

// 推荐: any (更简洁)
```

---

### 39. unsafe.Pointer
```go
// unsafe.Pointer 可以绕过类型系统
// 使用时必须非常小心

// 1. *T ↔ unsafe.Pointer
var x int = 42
p := unsafe.Pointer(&x)
px := (*int)(p)

// 2. unsafe.Pointer ↔ uintptr
ptr := uintptr(p)
ptr += 8
p2 := unsafe.Pointer(ptr)

// ⚠️ uintptr 没有引用语义，可能被 GC
// 必须确保在转换期间对象不会被 GC 移动
```

---

### 40. 内联优化
```go
// 简单函数会被内联
func add(a, b int) int {
    return a + b  // 可能内联
}

// 复杂函数不会被内联
// - 循环
// - 闭包
// - defer
// - 太大

// 禁止内联
//go:noinline
func noInline() {}
```

---

## 七、系统设计

### 41. 限流算法

**令牌桶:**
```go
// 恒定速率放入令牌，允许突发
// 实现: golang.org/x/time/rate
limiter := rate.NewLimiter(10, 100)  // 10/s, 容量100
limiter.Wait(ctx)  // 等待令牌
```

**漏桶:**
```go
// 恒定速率处理，平滑请求
// 适合消息队列
```

**固定窗口:**
```go
// 简单但不精确
// 边缘问题: 2倍流量
```

**滑动窗口:**
```go
// 更精确，内存占用大
// 记录每个请求时间戳
```

---

### 42. 分布式锁

**Redis:**
```go
// SET key value NX PX ttl
// 释放时检查 value 是否匹配

// 问题:
// - 时钟漂移
// - 主从切换导致锁丢失
```

**etcd:**
```go
// 事务 + lease
// 更可靠，但性能较低
```

**Zookeeper:**
```go
// 临时顺序节点
// 可靠性高，实现复杂
```

---

### 43. 幂等性设计
```go
// 1. 唯一ID
type Request struct {
    ID string  // 唯一标识
    // ...
}

// 2. 去重表
INSERT INTO dedup (id) VALUES (?)
ON CONFLICT (id) DO NOTHING

// 3. Token
token := generateToken()
// 客户端携带 token
// 服务端检查 token 是否已使用
```

---

### 44. 重试策略
```go
// 指数退避
func Retry(fn func() error, maxAttempts int) error {
    var err error
    for i := 0; i < maxAttempts; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        time.Sleep(time.Duration(1<<i) * time.Second)
    }
    return err
}

// 抖动: 避免惊群
delay := baseDelay * (1 + rand.Float64())
```

---

### 45. 熔断器
```go
type CircuitBreaker struct {
    maxFailures int
    resetTime   time.Duration
    // ...
}

// 状态: closed → open → half-open → closed
// closed: 正常
// open: 熔断，直接返回错误
// half-open: 尝试恢复
```

---

## 八、性能优化

### 46. pprof 使用
```bash
# CPU
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 内存
go test -memprofile=mem.prof
go tool pprof mem.prof

# goroutine
go test -goroutineprofile=goroutine.prof

# 火焰图
go tool pprof -http=:8080 cpu.prof
```

---

### 47. benchmark 技巧
```go
func BenchmarkFunc(b *testing.B) {
    b.ResetTimer()  // 重置计时器
    for i := 0; i < b.N; i++ {
        Func()
    }
}

// 并行 benchmark
func BenchmarkFuncParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Func()
        }
    })
}

// 子 benchmark
func BenchmarkFunc(b *testing.B) {
    b.Run("Case1", func(b *testing.B) { /*...*/ })
    b.Run("Case2", func(b *testing.B) { /*...*/ })
}
```

---

### 48. 内存对齐
```go
// 错误: 24 bytes
type Bad struct {
    a bool   // 1
    // 7 bytes padding
    b int64  // 8
    c bool   // 1
    // 7 bytes padding
}

// 正确: 16 bytes
type Good struct {
    b int64  // 8
    a bool   // 1
    c bool   // 1
    // 6 bytes padding
}

// 检查工具:
// fieldalignment
```

---

### 49. 字符串拼接
```go
// 慢: +
s := ""
for i := 0; i < 1000; i++ {
    s += "a"  // 每次分配新字符串
}

// 快: strings.Builder
var b strings.Builder
b.Grow(1000)
for i := 0; i < 1000; i++ {
    b.WriteString("a")
}
s := b.String()
```

---

### 50. slice vs array
```go
// array: 值传递，拷贝整个数组
func f(a [1000000]int) {
    // 拷贝 8MB
}

// slice: 引用传递，只拷贝头部
func f(s []int) {
    // 拷贝 24 字节
}

// 优化: 传指针
func f(a *[1000000]int) {
    // 拷贝 8 字节
}
```

---

## 快速参考: 命令速查

```bash
# 编译
go build              # 编译
go build -gcflags="-m" # 逃逸分析
go build -race        # 竞态检测

# 测试
go test -v            # 详细输出
go test -race         # 竞态检测
go test -cover        # 覆盖率
go test -bench=.      # 基准测试
go test -memprofile=mem.prof  # 内存分析

# 工具
go fmt ./...          # 格式化
go vet ./...          # 静态检查
golines ./...         # 长行检查
staticcheck ./...     # 高级检查

# 性能
go tool pprof          # pprof
go tool trace          # trace
go test -bench=. -benchmem  # 基准+内存
```
