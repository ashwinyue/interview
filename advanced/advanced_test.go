package advanced

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

// ============================================================================
// 第一部分：反射 (reflect)
// ============================================================================

// Q1: 反射的基本用法

func TestReflectionBasics(t *testing.T) {
	var x interface{} = 3.14

	// TypeOf 获取类型
	v := reflect.ValueOf(x)
	t.Logf("Type: %v", v.Type()) // float64
	t.Logf("Kind: %v", v.Kind()) // float64
	t.Logf("Value: %v", v)       // 3.14

	// 修改值需要传递指针
	x = 10
	v = reflect.ValueOf(&x)
	if v.CanSet() {
		t.Log("Can set directly")
	} else {
		t.Log("Cannot set, need Elem()")
	}

	v = v.Elem()
	if v.CanSet() {
		v.SetInt(20)
		t.Logf("After SetInt: x=%v", x) // x=20
	}
}

// Q2: 反射遍历结构体

type Person struct {
	Name string `json:"name" db:"person_name"`
	Age  int    `json:"age" db:"person_age"`
}

func (p Person) SayHello() {
	fmt.Printf("Hello, I'm %s\n", p.Name)
}

func TestReflectStruct(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}

	// 获取类型信息
	typ := reflect.TypeOf(p)
	t.Logf("Type name: %s", typ.Name())
	t.Logf("NumField: %d", typ.NumField())

	// 遍历字段
	val := reflect.ValueOf(p)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		t.Logf("Field %d: %s = %v (tag: %s)",
			i, field.Name, value.Interface(), field.Tag)
	}

	// 调用方法
	t.Logf("NumMethod: %d", typ.NumMethod())
	method, _ := typ.MethodByName("SayHello")
	t.Logf("Method: %s, type: %s", method.Name, method.Type)
}

// Q3: 反射获取和设置结构体字段

func TestReflectSetField(t *testing.T) {
	p := Person{Name: "Bob", Age: 25}

	// 通过反射修改字段
	val := reflect.ValueOf(&p).Elem()

	// 按名称查找字段
	nameField := val.FieldByName("Name")
	if nameField.IsValid() && nameField.CanSet() {
		nameField.SetString("Charlie")
	}

	// 按索引查找字段
	ageField := val.Field(1)
	if ageField.CanSet() {
		ageField.SetInt(35)
	}

	t.Logf("After modification: %+v", p)
}

// Q4: 反射调用方法

func TestReflectCallMethod(t *testing.T) {
	p := Person{Name: "David", Age: 40}

	val := reflect.ValueOf(p)

	// 调用无参数方法
	method := val.MethodByName("SayHello")
	if method.IsValid() {
		method.Call(nil)
	}

	// 调用有参数方法
	result := val.MethodByName("String").Call(nil)
	if len(result) > 0 {
		t.Logf("String() = %s", result[0])
	}
}

// Q5: 反射创建实例

func TestReflectNew(t *testing.T) {
	typ := reflect.TypeOf(Person{})

	// 创建新实例
	newValue := reflect.New(typ) // 返回 *Person
	elem := newValue.Elem()

	// 设置字段值
	elem.FieldByName("Name").SetString("Eve")
	elem.FieldByName("Age").SetInt(28)

	// 转换为原始类型
	p := newValue.Interface().(*Person)
	t.Logf("Created: %+v", p)
}

// Q6: 反射处理 slice 和 map

func TestReflectCollection(t *testing.T) {
	// 处理 slice
	slice := reflect.MakeSlice(reflect.TypeOf([]int{}), 0, 5)
	slice = reflect.Append(slice, reflect.ValueOf(1))
	slice = reflect.Append(slice, reflect.ValueOf(2))
	t.Logf("Slice: %v", slice.Interface())

	// 处理 map
	mapType := reflect.TypeOf(map[string]int{})
	mapValue := reflect.MakeMapWithSize(mapType, 10)
	mapValue.SetMapIndex(reflect.ValueOf("a"), reflect.ValueOf(1))
	mapValue.SetMapIndex(reflect.ValueOf("b"), reflect.ValueOf(2))
	t.Logf("Map: %v", mapValue.Interface())
}

// Q7: 实现通用的 JSON 序列化器（简化版）

func MarshalCustom(v interface{}) (string, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := "{"

	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)

			// 获取 json tag
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}

			if i > 0 {
				result += ","
			}
			result += fmt.Sprintf("\"%s\":%v", tag, fieldValue.Interface())
		}
	}

	result += "}"
	return result, nil
}

func TestCustomMarshal(t *testing.T) {
	p := Person{Name: "Frank", Age: 50}

	// 使用标准库
	stdJSON, _ := json.Marshal(p)
	t.Logf("Standard JSON: %s", stdJSON)

	// 使用自定义
	customJSON, _ := MarshalCustom(p)
	t.Logf("Custom JSON: %s", customJSON)
}

// Q8: 反射的性能问题

func BenchmarkReflection(b *testing.B) {
	p := Person{Name: "Test", Age: 30}

	b.Run("DirectAccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = p.Name
			_ = p.Age
		}
	})

	b.Run("Reflection", func(b *testing.B) {
		val := reflect.ValueOf(p)
		for i := 0; i < b.N; i++ {
			_ = val.FieldByName("Name").String()
			_ = val.Field(1).Int()
		}
	})

	// 反射比直接访问慢约 100 倍
	// 但在适当场景（如序列化、ORM）下是必要的
}

// ============================================================================
// 第二部分：unsafe 包
// ============================================================================

// Q9: unsafe.Pointer 基本用法

func TestUnsafePointer(t *testing.T) {
	// unsafe.Pointer 可以绕过类型系统
	// 使用时必须非常小心

	// 1. *T -> unsafe.Pointer
	x := int64(123)
	p := unsafe.Pointer(&x)

	// 2. unsafe.Pointer -> *T
	p2 := (*int64)(p)
	t.Logf("*p2 = %d", *p2) // 123

	// 3. unsafe.Pointer -> uintptr -> unsafe.Pointer
	// 警告: 这种转换不安全，中间可能发生 GC
	// 正确做法: 在一个语句中完成所有操作
	_ = unsafe.Pointer(uintptr(p) + 8) // 演示用，忽略警告
	t.Logf("After arithmetic, memory may contain anything")
}

// Q10: 字符串和 []byte 零拷贝转换

func TestStringByteConversion(t *testing.T) {
	s := "hello, world"

	// 方法1：常规转换（有拷贝）
	b1 := []byte(s)
	t.Logf("Regular conversion: %v, len=%d", b1, len(b1))

	// 方法2：unsafe 零拷贝转换（危险！）
	// 注意：修改 b2 会影响 s 的底层数据，可能导致程序崩溃
	b2 := unsafeStringToBytes(s)
	t.Logf("Unsafe conversion: %v, len=%d", b2, len(b2))

	// 方法3：反过来
	s2 := unsafeBytesToString(b1)
	t.Logf("Bytes to string: %s", s2)
}

// 零拷贝转换：string -> []byte
// ⚠️ 危险：修改返回的 slice 会导致程序崩溃
func unsafeStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// 零拷贝转换：[]byte -> string
// ⚠️ 危险：修改原 slice 会影响 string
func unsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Q11: 结构体字段偏移

func TestStructOffset(t *testing.T) {
	type MyStruct struct {
		A bool  // offset: 0
		B int64 // offset: 8 (对齐)
		C bool  // offset: 16
	}

	s := MyStruct{A: true, B: 42, C: false}

	// 获取字段偏移
	offsetB := unsafe.Offsetof(s.B)
	t.Logf("Offset of B: %d", offsetB) // 8

	// 直接访问字段内存
	bPtr := (*int64)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + offsetB))
	t.Logf("B value via unsafe: %d", *bPtr) // 42
}

// Q12: atomic.Value 底层原理

// atomic.Value 内部使用 unsafe 存储任意类型
// 这里演示简化版

type AnyValue struct {
	ptr unsafe.Pointer
	typ unsafe.Pointer
}

func (v *AnyValue) Store(x interface{}) {
	// 获取 x 的类型和指针
	rv := reflect.ValueOf(x)
	v.ptr = unsafe.Pointer(rv.UnsafeAddr())
	v.typ = unsafe.Pointer(rv.Pointer())
}

func (v *AnyValue) Load() interface{} {
	// 简化版，实际实现更复杂
	if v.ptr == nil {
		return nil
	}
	// 需要根据类型还原
	return nil
}

// Q13: unsafe 的使用场景

func TestUnsafeUseCases(t *testing.T) {
	// 1. 性能优化（避免拷贝）

	// 2. 与 C 代码交互 (cgo)

	// 3. 访问私有字段（不推荐）
	// type Secret struct {
	// 	privateField string
	// }
	// s := Secret{privateField: "secret"}
	// field := unsafe.Pointer(&s)
	// ...

	// 4. 实现自定义同步原语

	// 但要记住：
	// - unsafe 代码的可移植性不保证
	// - 可能被 GC 移动对象
	// - 类型安全性被破坏
}

// ============================================================================
// 第三部分：泛型 (Go 1.18+)
// ============================================================================

// Q14: 泛型函数

// Min 返回两个值中的较小值
func Min[T ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// ordered 约束：支持比较的类型
type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

func TestGenericMin(t *testing.T) {
	t.Logf("Min(int): %d", Min(10, 5))          // 5
	t.Logf("Min(float64): %f", Min(3.14, 2.71)) // 2.71
	t.Logf("Min(string): %s", Min("a", "b"))    // "a"
}

// Q15: 泛型类型

// Stack 泛型栈
type Stack[T any] struct {
	elements []T
}

func (s *Stack[T]) Push(v T) {
	s.elements = append(s.elements, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.elements) == 0 {
		return zero, false
	}
	idx := len(s.elements) - 1
	v := s.elements[idx]
	s.elements = s.elements[:idx]
	return v, true
}

func TestGenericStack(t *testing.T) {
	// int 栈
	intStack := &Stack[int]{}
	intStack.Push(1)
	intStack.Push(2)
	v, _ := intStack.Pop()
	t.Logf("Popped: %d", v) // 2

	// string 栈
	strStack := &Stack[string]{}
	strStack.Push("hello")
	v2, _ := strStack.Pop()
	t.Logf("Popped: %s", v2) // "hello"
}

// Q16: 多个类型参数

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := MapKeys(m)
	t.Logf("Keys: %v", keys)
}

// Q17: 类型推断

func TestTypeInference(t *testing.T) {
	// 完全推断
	result1 := Min(10, 5) // T 推断为 int

	// 部分指定
	result2 := Min[int](10, 5)

	t.Logf("result1=%d, result2=%d", result1, result2)
}

// Q18: any vs interface{}

func processAny(v any) {
	// any 是 interface{} 的别名
	_ = v
}

func processInterface(v interface{}) {
	_ = v
}

func TestAnyVsInterface(t *testing.T) {
	// any 和 interface{} 完全等价
	// any 是 Go 1.18 引入的别名，更简洁

	processAny("hello")
	processInterface("world")

	t.Log("any and interface{} are the same")
}

// Q19: ~ (近似类型) 约束

type MyInt int

// 使用 ~ 可以接受底层类型相同但类型不同的值
func Double[T ~int](x T) T {
	return x * 2
}

func TestTildeConstraint(t *testing.T) {
	var i int = 10
	var mi MyInt = 20

	// 如果约束是 int (不带 ~)，MyInt 不匹配
	// 但带 ~ 后，MyInt 可以匹配

	t.Logf("Double(int): %d", Double(i))    // 20
	t.Logf("Double(MyInt): %d", Double(mi)) // 40
}

// Q20: comparable 约束

// Find 在 slice 中查找值
func Find[T comparable](slice []T, target T) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

func TestComparable(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	idx := Find(nums, 3)
	t.Logf("Found 3 at index: %d", idx) // 2
}

// comparable 是内置约束，表示可以用 == 和 != 比较
// 不能用于 slice、map、func 等

// ============================================================================
// 第四部分：Go 1.21+ 新特性
// ============================================================================

// Q21: slices 包

func TestSlicesPackage(t *testing.T) {
	// slices 包提供了通用的切片操作

	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}

	// BinarySearch - 二分查找（需要已排序）
	// target := 5
	// i, found := slices.BinarySearch(nums, target)

	// Contains - 是否包含
	// contains := slices.Contains(nums, 5)

	// Index - 查找索引
	// idx := slices.Index(nums, 4)

	// Sort - 排序
	// slices.Sort(nums)

	// Reverse - 反转
	// slices.Reverse(nums)

	// Clone - 克隆
	// cloned := slices.Clone(nums)

	// Compact - 去重相邻元素
	// compacted := slices.Compact(nums)

	t.Logf("Original: %v", nums)
}

// Q22: maps 包

func TestMapsPackage(t *testing.T) {
	// maps 包提供了通用的 map 操作

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	// Keys - 获取所有键
	// keys := maps.Keys(m)

	// Values - 获取所有值
	// values := maps.Values(m)

	// Equal - 比较两个 map 是否相等
	// equal := maps.Equal(m, map[string]int{"a": 1, "b": 2, "c": 3})

	// Clone - 克隆
	// cloned := maps.Clone(m)

	// Copy - 复制
	// dest := make(map[string]int, len(m))
	// maps.Copy(dest, m)

	t.Logf("Map: %v", m)
}

// Q23: log/slog 结构化日志

func TestSlog(t *testing.T) {
	// Go 1.21 引入的结构化日志
	// import "log/slog"

	// 基本用法
	// slog.Info("user login", "user_id", 123, "ip", "192.168.1.1")

	// 结构化
	// slog.LogAttrs(nil, slog.LevelInfo, "user login",
	//     slog.Int("user_id", 123),
	//     slog.String("ip", "192.168.1.1"),
	// )

	t.Log("slog requires Go 1.21+")
}

// Q24: min/max 内置函数

func TestBuiltinMinMax(t *testing.T) {
	// Go 1.21+ 新增内置函数

	t.Logf("min(1, 2, 3) = %d", min(1, 2, 3))             // 1
	t.Logf("max(1.5, 2.5, 3.5) = %f", max(1.5, 2.5, 3.5)) // 3.5

	// 支持任意可比较类型
	t.Logf("min(\"a\", \"b\") = %s", min("a", "b"))

	// 注意：与泛型 Min 函数冲突时，需要重命名
}

// Q25: clear 内置函数

func TestBuiltinClear(t *testing.T) {
	// Go 1.21+ 新增内置函数

	// 清空 slice
	s := []int{1, 2, 3}
	clear(s)
	t.Logf("After clear slice: len=%d, cap=%d, %v", len(s), cap(s), s)

	// 清空 map
	m := map[string]int{"a": 1, "b": 2}
	clear(m)
	t.Logf("After clear map: len=%d, %v", len(m), m)
}

// ============================================================================
// 辅助类型和函数
// ============================================================================

// String 实现 String() 方法
func (p Person) String() string {
	return fmt.Sprintf("%s (age %d)", p.Name, p.Age)
}
