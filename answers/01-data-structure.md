# Go 数据结构面试原理详解

## 1. Slice 底层原理

### Q: 详细讲解 slice 的底层数据结构和扩容机制？

#### A: Slice 底层结构

```go
type slice struct {
    ptr unsafe.Pointer  // 指向底层数组的指针
    len int             // 当前长度
    cap int             // 容量
}
```

**三个关键字段：**
- `ptr`: 指向底层数组首元素的指针
- `len`: slice 中当前元素个数
- `cap`: 底层数组能容纳的元素个数

**核心要点：**
1. Slice 本身是值类型（但包含指针，本质是引用语义）
2. 多个 slice 可以共享同一个底层数组
3. 切片操作 `slice[start:end]` 不创建新数组，只是创建了新的 slice 结构

---

#### 扩容机制

Go 1.18 之后的扩容策略：

```go
func growslice(et *_type, old slice, nc int) slice {
    newcap := old.cap
    doublecap := newcap + newcap

    if nc > doublecap {
        newcap = nc
    } else {
        const threshold = 256
        if old.cap < threshold {
            newcap = doublecap  // 小于256，直接翻倍
        } else {
            // 大于256，增长约 25% (实际是 newcap += (newcap >> 3))
            for 0 < newcap && newcap < nc {
                newcap += (newcap >> 3) // +12.5%
            }
            if newcap <= 0 {
                newcap = nc
            }
        }
    }
    // 内存对齐调整...
}
```

**扩容规则总结：**

| 原容量 | 期望容量 | 新容量策略 |
|--------|----------|------------|
| < 256 | 翻倍 | `newcap = old.cap * 2` |
| ≥ 256 | 增长 ~25% | `newcap += newcap >> 3` |
| 任意 | 超过翻倍 | `newcap = nc` |

**面试常考点：**
```go
s := make([]int, 0, 1)
s = append(s, 1)   // len=1, cap=1
s = append(s, 2)   // len=2, cap=2 (翻倍)
s = append(s, 3)   // len=3, cap=4 (翻倍)
// ... 继续到 256 后，增长变慢
```

---

#### Slice 共享陷阱

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]  // b = [2, 3]

b[0] = 999
// 此时 a = [1, 999, 3, 4, 5] 修改 b 影响了 a！

b = append(b, 100, 200)
// 如果 cap 不够，b 会扩容并指向新数组
// 如果 cap 够，修改会影响 a！
```

**避免共享问题的方法：**
```go
// 方法1: 使用 copy
b := make([]int, len(a[1:3]))
copy(b, a[1:3])

// 方法2: 直接 append 创建新 slice
b := append([]int{}, a[1:3]...)
```

---

## 2. Map 底层原理

### Q: Go map 的底层实现是怎样的？为什么并发读写会 panic？

#### A: Map 底层结构

```go
type hmap struct {
    count     int    // 元素个数
    flags     uint8  // 状态标志
    B         uint8  // 桶数组大小的对数，桶数=2^B
    noverflow uint16 // 溢出桶数量
    hash0     uint32 // hash 种子

    buckets   unsafe.Pointer  // 桶数组指针
    oldbuckets unsafe.Pointer // 扩容时的旧桶数组
    nevacuate uintptr         // 搬迁进度
}

type bmap struct {
    tophash [bucketCnt]uint8  // 8个key的hash高8位
    // 后面是8个key、8个value、1个溢出指针
}
```

**核心设计：**
1. **桶结构**: 每个桶可存 8 个 key-value
2. **Hash 计算**: `hash(key) + hash0` → 取低 B 位 → 桶索引
3. **TopHash**: 存储 hash 的高 8 位，用于快速比较
4. **溢出**: 链表法处理冲突

**查找过程：**
```
1. 计算哈希: hash = hash_func(key, hash0)
2. 定位桶: index = hash & (1<<B - 1)
3. 遍历桶: 比较 tophash，匹配则返回
4. 溢出桶: 未找到则遍历溢出链表
```

---

#### 为什么并发读写会 panic？

Go map 的设计是**非线程安全**的，原因：

1. **性能考虑**: 大多数场景是单线程或已加锁
2. **避免误用**: 强制开发者显式处理并发

```go
var m = make(map[int]int)

// 并发写 - panic
go func() { m[1] = 1 }()
go func() { m[2] = 2 }()

// 并发读写 - panic
go func() { _ = m[1] }()
go func() { m[1] = 1 }()
```

**并发写检测原理：**
```go
type hmap struct {
    flags uint8  // 包含 iterator=1, oldIterator=2
}

// 写操作时检查
if h.flags&hashWriting != 0 {
    throw("concurrent map writes")
}
h.flags |= hashWriting
// ... 写操作
h.flags &^= hashWriting
```

**解决方案：**

| 方案 | 适用场景 | 性能 |
|------|----------|------|
| `sync.Mutex` | 通用 | 中 |
| `sync.RWMutex` | 读多写少 | 高读性能 |
| `sync.Map` | 读多写少、key 稳定 | 特定场景快 |

**sync.Map 原理：**
- `read`: `atomic.Value` 存的 readOnly map（无锁读）
- `dirty`: 普通 map（写操作）
- `misses`: read 未命中计数，达到阈值提升 dirty

---

#### Map 扩容

**触发条件：**
```go
// 负载因子 > 6.5 (元素数 / 桶数)
if count > bucketCnt*float64(len(buckets)) {
    hashGrow()
}
```

**扩容类型：**
1. **增量扩容**: 桶数翻倍，重新分配
2. **等量扩容**: 溢出桶太多，重新排列

**搬迁是渐进的：**
- 写操作时搬迁 1-2 个桶
- 读操作时搬迁当前桶
- 避免一次性搬迁的性能抖动

---

## 3. Interface 底层原理

### Q: Go 接口的底层实现是什么？`interface{}` == nil 的陷阱？

#### A: Interface 底层结构

```go
type iface struct {
    tab  *itab    // 方法表
    data unsafe.Pointer  // 数据指针
}

type eface struct {  // 空接口 interface{}
    _type *_type   // 类型信息
    data unsafe.Pointer  // 数据指针
}

type itab struct {
    inter *interfacetype  // 接口类型
    _type *_type          // 具体类型
    hash  uint32          // 类型 hash
    _     [4]byte
    fun   [1]uintptr      // 方法函数指针数组
}
```

**关键理解：**
1. 普通接口: `(type, value)` + 方法表
2. 空接口: 只有 `(type, value)`
3. 接口变量包含两个指针，共 16 字节（64位）

---

#### interface{} == nil 的陷阱

```go
func test() {
    var i interface{} = nil
    fmt.Println(i == nil)  // true

    var p *int = nil
    i = p
    fmt.Println(i == nil)  // false! 陷阱！

    // i 的内部结构:
    // type = *int (非 nil)
    // data = nil
    // 所以 i != nil
}
```

**判断接口是否真正为 nil：**
```go
func IsNil(i interface{}) bool {
    if i == nil {
        return true
    }
    // 使用反射判断 data 是否为 nil
    v := reflect.ValueOf(i)
    return v.IsNil()
}
```

---

#### 类型断言与 Type Switch

```go
// 类型断言 - 安全方式
if v, ok := i.(int); ok {
    fmt.Println("int:", v)
}

// Type Switch
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown:", v)
}
```

**性能考虑：**
- 类型断言比较快（直接比较类型指针）
- 反射慢（涉及字符串比较）
- 热路径避免反射

---

## 4. Defer 底层原理

### Q: defer 的执行顺序和性能开销？

#### A: Defer 底层结构

```go
type _defer struct {
    siz     int32   // 参数和返回值大小
    started bool
    heap    bool    // 是否分配在堆上
    sp      uintptr // 栈指针
    pc      uintptr // 程序计数器
    fn      *funcval // 延迟函数
    _panic  *_panic // 关联的 panic
    link    *_defer // 链表下一个
}
```

**执行顺序：LIFO（后进先出）**

```go
func main() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// 输出: 3 2 1
```

**defer 参数求值时机：**
```go
i := 0
defer func(n int) {
    fmt.Println(n)  // 输出 0，不是 1
}(i)  // 参数立即求值
i = 1
```

---

#### 性能优化

**Go 1.14+ 优化：** 开放编码（Open-coded）

符合条件的 defer（最多 8 个、无循环、不捕获外层变量）
会被内联，性能接近直接调用。

**优化建议：**
```go
// 不推荐：循环内 defer
for _, file := range files {
    defer file.Close()  // 可能内存泄漏
}

// 推荐：使用函数封装
func process(file *File) {
    defer file.Close()
    // 处理逻辑
}
for _, file := range files {
    process(file)
}
```

---

#### defer + 命名返回值

```go
func fn() (result int) {
    defer func() {
        result++  // 可以修改返回值
    }()
    return 0  // 返回 1
}
```

**原理：** 命名返回值在函数栈上分配，
defer 闭包捕获的是同一块内存。

---

## 5. String 与 []byte 转换

### Q: string 和 []byte 转换的内存开销？

#### A: 转换原理

```go
// string 结构
type stringStruct struct {
    str unsafe.Pointer  // 字符数组指针
    len int
}

// slice 结构
type slice struct {
    array unsafe.Pointer
    len   int
    cap   int
}
```

**常规转换（有拷贝）：**
```go
s := "hello"
b := []byte(s)  // 分配新内存，拷贝数据
```

**零拷贝转换（unsafe，危险）：**
```go
// string -> []byte
func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(
        &struct {
            string
            int
        }{s, len(s)},
    ))
}

// []byte -> string
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
```

**风险：** 如果后续修改 slice，string 不可变性被破坏，
可能导致程序崩溃。

---

#### string 不可变性

```go
s := "hello"
// s[0] = 'H'  // 编译错误

b := []byte(s)
b[0] = 'H'  // OK，不影响 s
```

**不可变的好处：**
1. 字符串可以安全地共享底层内存
2. 作为 map key 时无需拷贝
3. 并发安全（只读）
