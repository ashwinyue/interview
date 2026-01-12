package memory

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// ============================================================================
// 第一部分：栈与堆分配
// ============================================================================

// Q1: 理解栈分配 vs 堆分配
// 栈：快、自动回收、有限大小
// 堆：慢、GC 回收、大内存

func TestStackVsHeap(t *testing.T) {
	// 栈分配：变量不逃逸
	a := 1
	b := 2
	c := add(a, b)
	t.Logf("栈分配: a=%d, b=%d, c=%d", a, b, c)

	// 堆分配：变量逃逸
	s := make([]int, 1000)
	t.Logf("堆分配: slice len=%d, cap=%d", len(s), cap(s))
}

func add(a, b int) int {
	return a + b // a, b, c 都在栈上
}

// Q2: 逃逸分析
// 使用: go build -gcflags="-m" 查看逃逸分析结果

func TestEscapeAnalysis(t *testing.T) {
	// 情况1: 返回局部变量的指针 - 逃逸到堆
	p1 := newInt()
	t.Logf("逃逸到堆: *p1=%d, 地址=%p", *p1, p1)

	// 情况2: 指针赋值给接口 - 逃逸
	var i interface{} = newInt()
	t.Logf("接口逃逸: i=%v", i)

	// 情况3: 传给 interface{} 参数
	someFunction(newInt())

	// 情况4: slice 容量过大可能逃逸
	large := make([]int, 10000)
	t.Logf("大 slice 可能逃逸: len=%d", len(large))

	// 情况5: 闭包捕获变量 - 逃逸
	f := closureCapture(10)
	t.Logf("闭包捕获: f()=%d", f())
}

// newInt 返回 *int，变量会逃逸到堆
//
//go:noinline
func newInt() *int {
	i := 42
	return &i // i 逃逸到堆
}

func someFunction(i interface{}) {
	fmt.Println(i)
}

func closureCapture(x int) func() int {
	return func() int {
		return x + 1 // x 被闭包捕获，逃逸
	}
}

// Q3: 避免逃逸的优化技巧

// 优化前：逃逸
func createUserBad() *User {
	u := User{Name: "bad"}
	return &u // u 逃逸
}

// 优化后：不逃逸（调用者分配）
func createUserGood(u *User) {
	u.Name = "good"
}

type User struct {
	Name string
	Age  int
}

func TestAvoidEscape(t *testing.T) {
	// 优化前：堆分配
	u1 := createUserBad()
	t.Logf("逃逸: %+v", u1)

	// 优化后：栈分配
	var u2 User
	createUserGood(&u2)
	t.Logf("不逃逸: %+v", u2)
}

// ============================================================================
// 第二部分：GC (垃圾回收)
// ============================================================================

// Q4: 理解 Go GC 触发条件

func TestGCTrigger(t *testing.T) {
	// 触发条件：
	// 1. 分配内存时， heap-live > heap-mark * (1 + gc-percent/100)
	// 2. 定时触发 (forcegcperiod = 2分钟)
	// 3. 手动触发 runtime.GC()

	// 设置 GOGC 环境变量控制 GC 频率
	// GOGC=off 关闭 GC
	// GOGC=100 默认值，heap 增长 100% 后触发

	runtime.GC() // 手动触发 GC
	t.Log("手动触发 GC")

	// 查看 GC 统计信息
	var gcStats runtime.MemStats
	runtime.ReadMemStats(&gcStats)
	t.Logf("GC 次数: %d", gcStats.NumGC)
	t.Logf("暂停时间(纳秒): %d", gcStats.PauseTotalNs)
	t.Logf("上次 GC 暂停时间: %v", gcStats.LastGC)
}

// Q5: 三色标记法理解

/*
三色标记法过程：
1. 初始：所有对象都是白色
2. 标记：从根对象开始遍历
   - 灰色：已发现但未处理完引用的对象
   - 黑色：已处理完引用的对象
   - 白色：未发现的对象
3. 清理：所有白色对象可回收

写屏障 (Write Barrier)：
- 在 GC 期间，当修改指针时，通过写屏障保证一致性
- 插入写屏障：记录新引用
- 删除写屏障：记录被删除的引用
*/

// Q6: 常见内存泄漏场景

func TestMemoryLeaks(t *testing.T) {
	// 泄漏1: 子字符串泄漏
	// 大字符串：s = "这是一个非常非常长的字符串..."
	// 子串：sub = s[0:2] = "这是"
	// 问题：sub 仍然引用整个大字符串，导致大字符串无法回收

	largeString := make([]byte, 1024*1024) // 1MB
	for i := range largeString {
		largeString[i] = 'a'
	}

	// 泄漏方式：sub 引用整个 largeString
	sub := largeString[:10]
	_ = sub // largeString 整个 1MB 无法回收

	// 正确做法：复制到新的 slice
	subCopy := make([]byte, 10)
	copy(subCopy, largeString[:10])
	// 现在 largeString 可以被 GC 回收
	t.Logf("正确做法: 复制子串 len=%d", len(subCopy))

	// 泄漏2: 切片截断后的底层数组泄漏
	// 问题：切片虽然 len=0，但 cap 仍然很大
	data := make([]int, 1000000)
	data = data[:0] // len=0, 但 cap=1000000，底层数组不释放

	// 正确做法：
	data = nil // 或者 data = make([]int, 0, 小容量)

	// 泄漏3: 闭包捕获变量
	// 见 TestEscapeAnalysis 中的闭包示例

	// 泄漏4: goroutine 泄漏
	// 见 concurrency/patterns_test.go 中的示例

	// 泄漏5: time.After 泄漏
	// time.After 创建的 timer 在 channel 读取前不会被回收
	// 正确做法：使用 time.NewTimer 并调用 Stop()

	// 错误：
	// select {
	// case <-ch:
	//     return
	// case <-time.After(time.Minute): // timer 泄漏
	// }

	// 正确：
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	select {
	case <-make(chan struct{}):
		return
	case <-timer.C:
	}

	// 泄漏6: sync.Pool 长期持有引用
	// 如果 Put 进 Pool 的对象很大且不再使用，会影响 GC
	// 解决：定期清理或使用弱引用（Go 没有内置弱引用）
}

// Q7: map 删除后内存释放

func TestMapDelete(t *testing.T) {
	// Go 1.20+: map 删除后会逐步释放内存
	// Go 1.20 之前：删除不会立即释放内存

	m := make(map[int][]byte)

	// 添加大量数据
	for i := 0; i < 10000; i++ {
		m[i] = make([]byte, 1024)
	}

	runtime.GC()
	var stats1 runtime.MemStats
	runtime.ReadMemStats(&stats1)

	// 删除所有数据
	for i := range m {
		delete(m, i)
	}

	runtime.GC()
	var stats2 runtime.MemStats
	runtime.ReadMemStats(&stats2)

	t.Logf("删除前堆内存: %d MB", stats1.HeapAlloc/1024/1024)
	t.Logf("删除后堆内存: %d MB", stats2.HeapAlloc/1024/1024)

	// 强制 map 释放内存（Go 1.20+）
	m = nil
	runtime.GC()

	var stats3 runtime.MemStats
	runtime.ReadMemStats(&stats3)
	t.Logf("置为 nil 后堆内存: %d MB", stats3.HeapAlloc/1024/1024)
}

// ============================================================================
// 第三部分：内存优化技巧
// ============================================================================

// Q8: 减少小对象分配

// 优化前：每个节点单独分配
type NodeBad struct {
	value int
	next  *NodeBad
}

// 优化后：使用 slice 存储，连续内存
type NodesGood struct {
	values []int
	nexts  []int // 索引指向下一个节点
}

func TestMemoryLayout(t *testing.T) {
	// 查看结构体内存布局
	type A struct {
		a bool  // 1 byte
		b int64 // 8 bytes
		c bool  // 1 byte
		// 实际占用: 1 + 7(padding) + 8 + 1 + 7(padding) = 24 bytes
	}

	type B struct {
		b int64 // 8 bytes
		a bool  // 1 byte
		c bool  // 1 byte
		// 实际占用: 8 + 1 + 1 + 6(padding) = 16 bytes
	}

	t.Logf("A size: %d bytes", unsafe.Sizeof(A{})) // 24
	t.Logf("B size: %d bytes", unsafe.Sizeof(B{})) // 16

	// 优化原则：按大小递减排序字段
}

// Q9: 使用 sync.Pool 复用对象

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func getBuffer() []byte {
	return bufferPool.Get().([]byte)
}

func putBuffer(b []byte) {
	// 重置后放回
	b = b[:0]
	bufferPool.Put(b)
}

func TestSyncPool(t *testing.T) {
	buf := getBuffer()
	buf = append(buf, "hello"...)
	putBuffer(buf)

	// sync.Pool 的特点：
	// 1. 每个 P 有自己的本地缓存
	// 2. GC 时会清除 Pool 中的对象
	// 3. 适合临时对象复用，减少 GC 压力
}

// Q10: 字符串拼接优化

func TestStringConcat(t *testing.T) {
	// 方法1: + 运算符 (每次都创建新字符串)
	s1 := ""
	for i := 0; i < 100; i++ {
		s1 += "a" // 每次分配新字符串
	}

	// 方法2: strings.Builder (推荐)
	var b2 StringBuilder
	b2.Grow(100)
	for i := 0; i < 100; i++ {
		b2.WriteString("a")
	}
	s2 := b2.String()

	// 方法3: []byte + 转换
	bs := make([]byte, 0, 100)
	for i := 0; i < 100; i++ {
		bs = append(bs, 'a')
	}
	s3 := string(bs)

	t.Logf("s1 len=%d, s2 len=%d, s3 len=%d", len(s1), len(s2), len(s3))
}

// Q11: 切片预分配

func TestSlicePreallocation(t *testing.T) {
	// 优化前：多次扩容
	var s1 []int
	for i := 0; i < 1000000; i++ {
		s1 = append(s1, i) // 多次扩容和复制
	}

	// 优化后：预分配
	s2 := make([]int, 0, 1000000)
	for i := 0; i < 1000000; i++ {
		s2 = append(s2, i) // 无扩容
	}

	t.Logf("s1 cap=%d, s2 cap=%d", cap(s1), cap(s2))
}

// Q12: 减少指针引用

func TestPointerFree(t *testing.T) {
	// 指针会阻止 GC 的并发扫描
	// 尽量使用值类型而非指针类型

	// 优化前：slice of pointers
	type PointPtr struct {
		X *int
		Y *int
	}
	// 存储时会有额外指针，增加 GC 扫描时间

	// 优化后：slice of values
	type PointValue struct {
		X int
		Y int
	}
	// 连续内存，减少 GC 压力

	t.Logf("PointPtr size: %d, PointValue size: %d",
		unsafe.Sizeof(PointPtr{}), unsafe.Sizeof(PointValue{}))
}

// ============================================================================
// 第四部分：性能分析工具
// ============================================================================

func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("NoPreallocation", func(b *testing.B) {
		var s []int
		for i := 0; i < b.N; i++ {
			s = append(s, i)
		}
	})

	b.Run("WithPreallocation", func(b *testing.B) {
		s := make([]int, 0, b.N)
		for i := 0; i < b.N; i++ {
			s = append(s, i)
		}
	})
}

// 内存分析命令：
// go test -bench=. -benchmem -memprofile=mem.prof
// go tool pprof mem.prof
// (pprof) top10
// (pprof) list 函数名

// CPU 分析命令：
// go test -bench=. -cpuprofile=cpu.prof
// go tool pprof cpu.prof

// 逃逸分析命令：
// go build -gcflags="-m" main.go
// go test -gcflags="-m" ./...

// ============================================================================
// 第五部分：内存对齐
// ============================================================================

func TestMemoryAlignment(t *testing.T) {
	// CPU 访问对齐的内存更快
	// 64位系统：8字节对齐
	// 32位系统：4字节对齐

	type Bad struct {
		a bool // offset: 0, size: 1
		// padding: 7 bytes
		b int64 // offset: 8, size: 8
		c bool  // offset: 16, size: 1
		// padding: 7 bytes
		// total: 24 bytes
	}

	type Good struct {
		b int64 // offset: 0, size: 8
		a bool  // offset: 8, size: 1
		c bool  // offset: 9, size: 1
		// padding: 6 bytes
		// total: 16 bytes
	}

	// 使用 fieldalignment 工具检查：
	// go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	// fieldalignment ./...

	t.Logf("Bad: %d bytes, Good: %d bytes", unsafe.Sizeof(Bad{}), unsafe.Sizeof(Good{}))
}

// ============================================================================
// 辅助函数
// ============================================================================

// StringBuilder 模拟 strings.Builder
type StringBuilder struct {
	buf []byte
}

func (b *StringBuilder) Grow(n int) {
	if cap(b.buf)-len(b.buf) < n {
		newBuf := make([]byte, len(b.buf), len(b.buf)+n)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}

func (b *StringBuilder) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *StringBuilder) String() string {
	return string(b.buf)
}

// printMemStats 打印内存统计
func printMemStats(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Alloc = %v MB", m.Alloc/1024/1024)
	t.Logf("TotalAlloc = %v MB", m.TotalAlloc/1024/1024)
	t.Logf("Sys = %v MB", m.Sys/1024/1024)
	t.Logf("NumGC = %v", m.NumGC)
}
