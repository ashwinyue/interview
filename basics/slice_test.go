package basics

import "testing"

// 练习1：理解 slice 底层结构
// Slice 包含三个字段：指向底层数组的指针、长度、容量

func TestSliceBasics(t *testing.T) {
	// 1.1 nil slice vs empty slice
	var s1 []int          // nil slice
	s2 := []int{}        // empty slice
	s3 := make([]int, 0) // empty slice

	// TODO: 判断以下表达式的结果
	t.Logf("s1 == nil: %v", s1 == nil)
	t.Logf("s2 == nil: %v", s2 == nil)
	t.Logf("s3 == nil: %v", s3 == nil)
	t.Logf("len(s1): %d, cap(s1): %d", len(s1), cap(s1))
}

func TestSliceGrowth(t *testing.T) {
	// 1.2 slice 扩容机制
	s := make([]int, 0, 1)

	t.Logf("初始: len=%d, cap=%d", len(s), cap(s))

	s = append(s, 1)
	t.Logf("append 1 后: len=%d, cap=%d", len(s), cap(s))

	s = append(s, 2)
	t.Logf("append 2 后: len=%d, cap=%d", len(s), cap(s))

	// TODO: 观察扩容规律
	// Go 1.18+ 扩容策略：
	// - 期望容量 > 256: 新容量 = 旧容量 * 1.25
	// - 期望容量 <= 256: 新容量 = 旧容量 * 2 (可能有内存对齐调整)
}

func TestSliceSharing(t *testing.T) {
	// 1.3 slice 共享底层数组
	s1 := []int{1, 2, 3, 4, 5}
	s2 := s1[1:3] // s2 = [2, 3]

	t.Logf("s1: %v, s2: %v", s1, s2)

	s2[0] = 999 // 修改 s2 会影响 s1 吗？

	t.Logf("修改后 s1: %v, s2: %v", s1, s2)

	// TODO: 理解 slice 共享的潜在问题
	// 如何避免共享问题？使用 copy()
}

// 练习2：Map 线程安全
func TestMapConcurrency(t *testing.T) {
	// 2.1 多个 goroutine 并发读写 map 会 panic
	// TODO: 演示 map 的并发安全问题（会有 panic）
	// 解决方案：
	// 1. 使用 sync.Map
	// 2. 使用 map + sync.Mutex / sync.RWMutex
}

// 练习3：Interface 底层结构
// interface value 包含两个指针：(type, value)

type I interface {
	M()
}

type T struct {
	S string
}

func (t *T) M() {}

func TestInterfaceNil(t *testing.T) {
	var i I
	var ptr *T

	// TODO: 判断以下结果
	t.Logf("i == nil: %v", i == nil) // true

	i = ptr
	t.Logf("i == nil: %v", i == nil) // false! 为什么？
	// i 的 type 是 *T，value 是 nil

	// 如何判断接口值是否真正为 nil？
	if i == nil {
		t.Log("i is nil")
	} else if reflectValueIsNil(i) {
		t.Log("i holds a nil value")
	}
}

func reflectValueIsNil(i I) bool {
	// TODO: 使用反射判断接口值是否为 nil
	return false
}

// 练习4：Defer 执行顺序
func TestDeferOrder(t *testing.T) {
	// defer 按照后进先出 (LIFO) 顺序执行
	defer func() { t.Log("defer 1") }()
	defer func() { t.Log("defer 2") }()
	defer func() { t.Log("defer 3") }()

	// TODO: 输出顺序是什么？
}

func TestDeferParamEval(t *testing.T) {
	i := 0
	defer func(n int) { t.Logf("defer: n=%d", n) }(i) // 立即求值
	i++

	// TODO: 输出的是什么？0 还是 1？
}

// 练习5：闭包与循环变量
func TestClosureLoop(t *testing.T) {
	// Go 1.22 之前：循环变量是单个变量，每次迭代更新值
	// Go 1.22+：每次迭代创建新的变量

	var funcs []func()
	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() {
			t.Logf("i = %d", i)
		})
	}

	for _, f := range funcs {
		f()
	}

	// TODO: Go 1.22+ 输出什么？
	// 如果在 Go 1.22 之前会输出什么？
}
