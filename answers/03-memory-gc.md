# Go 内存管理与GC面试原理详解

## 1. 内存分配模型

### Q: Go 的内存分配策略是怎样的？TCMalloc 的借鉴？

#### A: Go 内存分配器概述

Go 采用基于 TCMalloc 的多级缓存分配器，核心思想：

```
┌─────────────────────────────────────────────────────┐
│                   mheap (全局堆)                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐             │
│  │ arena 0 │  │ arena 1 │  │ arena 2 │  ...         │
│  │ 512MB   │  │ 512MB   │  │ 512MB   │             │
│  └─────────┘  └─────────┘  └─────────┘             │
│                                                       │
│  每个 arena 分为:                                     │
│  ┌────────────────┬──────────────────────────────┐  │
│  │   spans        │    bitmap     │  arena (页)    │  │
│  │  (元数据)      │   (标记位)    │   (实际分配)    │  │
│  └────────────────┴──────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

#### mcache → mcentral → mheap

**三级缓存架构：**

```
                    ┌─────────────────┐
                    │     mheap       │ 全局
                    │  (mspan 列表)    │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
         ┌─────────┐   ┌─────────┐   ┌─────────┐
         │mcentral │   │mcentral │   │mcentral │ 全局
         │ span 0  │   │ span 1  │   │ span 2  │
         └────┬────┘   └────┬────┘   └────┬────┘
              │             │             │
         ┌────┴────┬────────┴────┬────────┴────┐
         ▼         ▼             ▼             ▼
      ┌───────┐ ┌───────┐   ┌───────┐   ┌───────┐
      │mcache │ │mcache │   │mcache │   │mcache │ 每个 P 一个
      │  P0   │ │  P1   │   │  P2   │   │  Pn   │
      └───────┘ └───────┘   └───────┘   └───────┘
         │
         │ mcache 内部 (size class 0~67)
         ▼
      ┌────┬────┬────┬────┬────┬────┐
      │ sc0│sc1 │sc2 │... │sc66│sc67│ 微对象 < 16B
      └────┴────┴────┴────┴────┴────┘
```

**分配路径：**

| 对象大小 | 分配路径 | 说明 |
|----------|----------|------|
| `< 16B` | mcache 微对象分配器 | 无需锁 |
| `< 32KB` | mcache → mcentral → mheap | mcache 命中无需锁 |
| `≥ 32KB` | mheap 直接分配 | 需要锁 |

**mspan (内存跨度):**
```go
type mspan struct {
    next     *mspan  // 链表下一个
    prev     *mspan  // 链表前一个
    startAddr uintptr // 起始地址
    npages    uintptr // 页数 (8KB/页)
    manualFreeList gclinkptr // 释放链表
    // ...
}
```

---

#### size class (大小类别)

Go 将对象大小分为 67 个 size class：

```
class 0:   0  B
class 1:   8  B
class 2:  16  B
class 3:  24  B
...
class 66: 28672 B (28KB)
class 67: 32768 B (32KB)
```

**分配策略：**
- 8 字节对齐
- 每个对应不同的 span
- 减少碎片，提高复用

---

## 2. 栈与堆分配

### Q: 变量什么时候分配在栈，什么时候分配在堆？

#### A: 逃逸分析 (Escape Analysis)

**逃逸分析编译器：**
```bash
go build -gcflags="-m" main.go
```

**逃逸场景：**

| 场景 | 示例 | 是否逃逸 |
|------|------|----------|
| 返回局部变量指针 | `return &local` | ✅ 逃逸 |
| 指针赋值给接口 | `var i interface{} = &p` | ✅ 逃逸 |
| 切片/-map 容量过大 | `make([]int, 10000)` | ✅ 可能逃逸 |
| 闭包捕获变量 | `func() { return x }` | ✅ 逃逸 |
| 发送到 channel | `ch <- &local` | ✅ 逃逸 |
| 动态大小 | `make([]int, n)` | ✅ 可能逃逸 |
| 局部值类型 | `x := 42` | ❌ 不逃逸 |
| 返回值类型 | `return x` | ❌ 不逃逸 |

**示例分析：**
```go
// 逃逸：返回指针
func f1() *int {
    x := 42
    return &x  // x 逃逸到堆
}

// 不逃逸：返回值
func f2() int {
    x := 42
    return x  // x 在栈上
}

// 逃逸：赋值给接口
func f3() interface{} {
    x := 42
    return x  // x 逃逸（接口是引用类型）
}

// 逃逸：闭包捕获
func f4() func() int {
    x := 42
    return func() int {
        return x  // x 被闭包捕获，逃逸
    }
}
```

---

#### 栈与堆的对比

| 特性 | 栈 | 堆 |
|------|-----|-----|
| 分配速度 | 极快 (SP 移动) | 较慢 (查找空闲块) |
| 释放速度 | 自动 (函数返回) | GC 回收 |
| 大小限制 | 小 (通常 2GB) | 大 (受物理内存) |
| 碎片问题 | 无 | 有 |
| 并发 | 无需锁 | 需要 |

**栈扩容：**
```go
// goroutine 初始栈: 2KB
// 扩容策略: 翻倍 (2KB → 4KB → 8KB ...)
// 最大栈: 1GB (但会检查)

func recursive(n int) {
    var buf [1024]int  // 8KB
    if n > 0 {
        recursive(n - 1)
    }
}
```

**连续栈 (Go 1.4+):**
- 旧栈拷贝到新栈
- 调整指针指向新位置
- 比分段栈性能更好

---

## 3. 垃圾回收 (GC)

### Q: Go GC 的三色标记法是怎样的？

#### A: GC 算法演进

```
Go 1.4  →  精确 GC
Go 1.5  →  并发三色标记 (STW ~10ms)
Go 1.6  →  支持 > 10GB 堆
Go 1.7  →  STW < 100μs
Go 1.8  →  混合写屏障
Go 1.9+ →  持续优化
```

---

#### 三色标记法详解

**三个颜色：**

```
白色 (White): 未访问的对象
灰色 (Gray):  已访问，但引用对象未全部访问
黑色 (Black): 已访问，引用对象也已全部访问
```

**标记过程：**

```
初始状态: 所有对象都是白色

┌─────┐  ┌─────┐  ┌─────┐
│  W  │  │  W  │  │  W  │  白色对象
└─────┘  └─────┘  └─────┘
   ▲
   │根对象 (全局变量、栈变量)
   ▼
┌─────┐  ┌─────┐  ┌─────┐
│  G  │→ │  W  │  │  W  │  G = 灰色
└─────┘  └─────┘  └─────┘
   ▼
┌─────┐  ┌─────┐  ┌─────┐
│  B  │→ │  G  │→ │  W  │  B = 黑色
└─────┘  └─────┘  └─────┘

最终: 黑色 = 活跃对象，白色 = 垃圾
```

**算法步骤：**
1. **标记准备**: STW，扫描根对象，标记为灰色
2. **并发标记**: 处理灰色对象，标记为黑色，新对象标记为灰色
3. **标记终止**: STW，重新扫描栈（可能遗漏的新引用）
4. **清理**: 并发清扫白色对象

---

#### 写屏障 (Write Barrier)

**问题：** 并发标记时，用户代码修改引用导致错误。

**示例场景：**
```
初始: A(黑) → B(灰) → C(白)

用户代码执行:
A.D = D  // 新建对象 D
B.C = nil // 删除引用

没有写屏障: C 被误认为垃圾回收！
```

**混合写屏障 (Go 1.8+):**
```go
// 伪代码
func writeBarrier(dst, src *object) {
    shade(dst)  // 将目标标灰
    shade(*dst) // 将原引用标灰
    *dst = src
}
```

**两种屏障：**
1. **插入写屏障**: 标记新引用
2. **删除写屏障**: 标记被删除的引用

Go 1.8 采用混合策略，结合两者优点。

---

#### GC 触发条件

```go
// 1. 堆内存增长阈值
if heap_marked > heap_goal {
    gcStart()
}

// heap_goal = heap_marked * (1 + GOGC/100)
// GOGC=100 默认值，堆增长 100% 触发
// GOGC=off 关闭 GC

// 2. 定时触发 (forcegcperiod)
// 每 2 分钟检查一次

// 3. 手动触发
runtime.GC()

// 4. 通过 runtime/debug 设置
debug.SetGCPercent(200)  // 相当于 GOGC=200
```

---

#### GC 性能优化

**减少 GC 压力：**
```go
// 1. 复用对象
var buf = make([]byte, 1024)

// 2. 使用 sync.Pool
pool := &sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

// 3. 避免指针
type Point struct {
    X, Y int  // 好
}
type Point struct {
    X, Y *int // 差，增加扫描时间
}

// 4. 预分配容量
s := make([]int, 0, 1000)  // 好
s := []int{}                // 差，多次扩容
```

---

## 4. 常见内存泄漏

### Q: Go 中常见的内存泄漏场景？

#### A: 内存泄漏场景总结

**1. Goroutine 泄漏**
```go
// 泄漏: goroutine 永远阻塞
func leak() {
    ch := make(chan int)
    go func() {
        <-ch  // 没有发送者，永远阻塞
    }()
}

// 修复
func fixed() {
    ch := make(chan int, 1)  // buffered
    // 或使用 context
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        select {
        case <-ch:
        case <-ctx.Done():
            return
        }
    }()
    cancel()
}
```

**2. 子字符串/切片泄漏**
```go
// 泄漏: 大字符串被小切片引用
large := strings.Repeat("a", 1024*1024)  // 1MB
small := large[:10]  // small 引用整个 large

// 修复: 复制
small := make([]byte, 10)
copy(small, large[:10])
```

**3. 切片底层数组泄漏**
```go
// 泄漏: 删除元素后底层数组仍持有
data := make([]int, 1000000)
data = data[:0]  // len=0, cap=1000000

// 修复
data = nil
// 或
data = make([]int, 0, small)
```

**4. 闭包泄漏**
```go
// 泄漏: 闭包捕获大对象
func f() {
    large := make([]byte, 1024*1024)
    go func() {
        // large 被闭包捕获，不能回收
        time.Sleep(time.Hour)
    }()
}

// 修复: 只捕获需要的
func f() {
    large := make([]byte, 1024*1024)
    needed := len(large)
    go func() {
        // 只捕获 needed
        time.Sleep(time.Hour)
    }()
}
```

**5. time.After 泄漏**
```go
// 泄漏: timer 未被读取
select {
case v := <-ch:
    return
case <-time.After(time.Minute):  // timer 泄漏
    return
}

// 修复: 使用 NewTimer
timer := time.NewTimer(time.Minute)
defer timer.Stop()
select {
case v := <-ch:
    return
case <-timer.C:
    return
}
```

**6. Finalizer 泄漏**
```go
// Finalizer 阻止对象回收
type File struct {
    fd int
}

func useFile() {
    f := &File{fd: open()}
    runtime.SetFinalizer(f, func(f *File) {
        close(f.fd)
    })
    // f 可能不会被回收
}
```

---

## 5. 内存分析工具

### Q: 如何分析和优化 Go 程序的内存？

#### A: 内存分析工具链

**1. 读取内存统计**
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)

fmt.Printf("Alloc = %v MB\n", m.Alloc/1024/1024)        // 当前分配
fmt.Printf("TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024) // 累计分配
fmt.Printf("Sys = %v MB\n", m.Sys/1024/1024)             // 从系统获取
fmt.Printf("NumGC = %v\n", m.NumGC)                      // GC 次数
fmt.Printf("NextGC = %v MB\n", m.NextGC/1024/1024)       // 下次 GC 触发点
```

**2. pprof 内存分析**
```bash
# 生成 profile
go test -memprofile=mem.prof
go tool pprof mem.prof

# 常用命令
(pprof) top10        # 前 10 内存分配
(pprof) list funcName  # 查看函数分配
(pprof) png          # 生成调用图
```

**3. trace 工具**
```bash
# 生成 trace
go test -trace=trace.out

# 可视化
go tool trace trace.out
```

**4. 逃逸分析**
```bash
# 查看逃逸分析
go build -gcflags="-m" main.go

# 更详细信息
go build -gcflags="-m -m" main.go
```

**5. GODEBUG 调试 GC**
```bash
# 查看 GC 信息
GODEBUG=gctrace=1 go run main.go

# 输出示例:
# gc 1 @0.001s 1%: 0.018+0.23+0.022 ms clock, 0.14+0.18/0.053+0.17 ms cpu, 4->6->4 MB, 5 MB goal, 12 P
```
