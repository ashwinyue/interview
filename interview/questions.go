package interview

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// 第一部分：Go 语言基础 (必考)
// ============================================================================

// Q1: 实现 LRU Cache (高频题)
// 要求：O(1) 时间复杂度获取和删除
type LRUCache struct {
	capacity int
	// TODO: 完成实现
}

func ConstructorLRU(capacity int) *LRUCache {
	return &LRUCache{capacity: capacity}
}

func (lru *LRUCache) Get(key int) int {
	// TODO: 实现
	return -1
}

func (lru *LRUCache) Put(key int, value int) {
	// TODO: 实现
}

// Q2: 实现两个 goroutine 交替打印数字和字母
// 输出: 1A2B3C4D... 或 A1B2C3D4...
func AlternatePrint() {
	// TODO: 使用 channel 或 sync.Cond 实现
}

// Q3: 实现一个线程安全的单例
type Singleton struct {
	value string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance() *Singleton {
	once.Do(func() {
		instance = &Singleton{value: "singleton"}
	})
	return instance
}

// Q4: 实现限流器 (令牌桶或漏桶)
type RateLimiter struct {
	rate     int       // 每秒令牌数
	capacity int       // 桶容量
	tokens   int       // 当前令牌数
	lastTime time.Time // 上次取令牌时间
	mu       sync.Mutex
}

func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		tokens:   capacity,
		lastTime: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()

	// 计算新增令牌数
	newTokens := int(elapsed * float64(rl.rate))
	if newTokens > 0 {
		rl.tokens = min(rl.capacity, rl.tokens+newTokens)
		rl.lastTime = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Q5: 实现分布式锁接口
type DistributedLock interface {
	Lock(ctx context.Context) error
	Unlock() error
	// 可选：续约、看门狗
}

// Q6: 实现一个简单的 TCP 服务端
// 要求：
// - 支持并发连接
// - 超时控制
// - 优雅关闭

// Q7: slice 去重
func RemoveDuplicates(nums []int) []int {
	// TODO: 实现多种方式
	// 1. 使用 map
	// 2. 排序后去重
	// 3. 双指针
	return nil
}

// Q8: 实现 merge k sorted lists
type ListNode struct {
	Val  int
	Next *ListNode
}

func MergeKLists(lists []*ListNode) *ListNode {
	// TODO: 使用堆 (container/heap) 实现
	return nil
}

// Q9: 实现一个简单的对象池
type ObjectPool struct {
	pool chan interface{}
	factory func() interface{}
}

func NewObjectPool(size int, factory func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool:    make(chan interface{}, size),
		factory: factory,
	}
}

func (p *ObjectPool) Get() interface{} {
	select {
	case obj := <-p.pool:
		return obj
	default:
		return p.factory()
	}
}

func (p *ObjectPool) Put(obj interface{}) {
	select {
	case p.pool <- obj:
	default:
		// 池满，丢弃
	}
}

// Q10: 实现一个带超时的 HTTP 请求
func RequestWithTimeout(url string, timeout time.Duration) error {
	// TODO: 使用 context.WithTimeout 实现
	return nil
}

// ============================================================================
// 第二部分：并发模式 (高频)
// ============================================================================

// Q11: 实现 Fan-Out/Fan-In
func FanOut(in <-chan int, workers int) []<-chan int {
	// TODO: 将输入分发到多个 worker
	return nil
}

func FanIn(channels ...<-chan int) <-chan int {
	// TODO: 聚合多个 channel 的输出
	return nil
}

// Q12: 实现一个安全的 map (sync.Map vs map+mutex)
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]interface{}),
	}
}

func (sm *SafeMap) Load(key string) (interface{}, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, ok := sm.data[key]
	return val, ok
}

func (sm *SafeMap) Store(key string, value interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

func (sm *SafeMap) LoadOrStore(key string, value interface{}) (interface{}, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if val, ok := sm.data[key]; ok {
		return val, true
	}
	sm.data[key] = value
	return value, false
}

func (sm *SafeMap) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

// Q13: 实现 Or-Done 模式
// 当任意一个 channel 关闭时，返回的 channel 也关闭
func OrDone(channels ...<-chan struct{}) <-chan struct{} {
	// TODO: 实现
	return nil
}

// Q14: 实现一个可取消的任务执行器
type TaskExecutor struct {
	ctx    context.Context
	cancel context.CancelFunc
	tasks  chan func()
	wg     sync.WaitGroup
}

func NewTaskExecutor(ctx context.Context, queueSize int) *TaskExecutor {
	childCtx, cancel := context.WithCancel(ctx)
	return &TaskExecutor{
		ctx:    childCtx,
		cancel: cancel,
		tasks:  make(chan func(), queueSize),
	}
}

func (e *TaskExecutor) Start(workers int) {
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.worker()
	}
}

func (e *TaskExecutor) worker() {
	defer e.wg.Done()
	for {
		select {
		case task := <-e.tasks:
			task()
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *TaskExecutor) Submit(task func()) error {
	select {
	case e.tasks <- task:
		return nil
	case <-e.ctx.Done():
		return e.ctx.Err()
	}
}

func (e *TaskExecutor) Stop() {
	e.cancel()
	e.wg.Wait()
}

// Q15: 实现一个带重试的函数
func RetryWithContext(ctx context.Context, fn func() error, maxAttempts int, backoff time.Duration) error {
	// TODO: 实现指数退避重试
	return nil
}

// ============================================================================
// 第三部分：算法题 (LeetCode 高频)
// ============================================================================

// Q16: 两数之和 - 返回索引
func TwoSum(nums []int, target int) []int {
	// TODO: 使用 map 实现 O(n) 时间复杂度
	return nil
}

// Q17: 反转链表
func ReverseList(head *ListNode) *ListNode {
	// TODO: 迭代和递归两种实现
	return nil
}

// Q18: 判断链表是否有环
func HasCycle(head *ListNode) bool {
	// TODO: 快慢指针
	return false
}

// Q19: 合并两个有序链表
func MergeTwoLists(l1, l2 *ListNode) *ListNode {
	// TODO: 递归和迭代两种实现
	return nil
}

// Q20: 二叉树层序遍历
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func LevelOrder(root *TreeNode) [][]int {
	// TODO: 使用队列实现
	return nil
}

// Q21: 快速排序
func QuickSort(nums []int) {
	// TODO: 实现原地快排
}

// Q22: 二分查找
func BinarySearch(nums []int, target int) int {
	// TODO: 迭代实现
	return -1
}

// Q23: 最长无重复子串
func LengthOfLongestSubstring(s string) int {
	// TODO: 滑动窗口
	return 0
}

// Q24: 爬楼梯 (动态规划)
func ClimbStairs(n int) int {
	// TODO: DP 或 矩阵快速幂
	return 0
}

// Q25: LRU 缓存 (见 Q1)

// ============================================================================
// 第四部分：系统设计题
// ============================================================================

// Q26: 设计一个短链接服务
// 需要考虑：
// - 如何生成短码？
// - 如何处理碰撞？
// - 如何存储？
// - 如何实现重定向？

// Q27: 设计一个秒杀系统
// 需要考虑：
// - 如何防止超卖？
// - 如何处理高并发？
// - 如何限流？
// - 如何削峰？

// Q28: 设计一个分布式 ID 生成器
// 需要考虑：
// - 全局唯一性
// - 趋势递增
// - 高可用

// Q29: 设计一个消息队列
// 需要考虑：
// - 消息可靠性
// - 顺序保证
// - 消息去重
// - 积压处理

// Q30: 设计一个分布式缓存
// 需要考虑：
// - 缓存穿透
// - 缓存击穿
// - 缓存雪崩
// - 一致性

// ============================================================================
// 辅助函数
// ============================================================================

// PrintSlice 打印 slice (调试用)
func PrintSlice(name string, s []int) {
	fmt.Printf("%s: len=%d cap=%d %v\n", name, len(s), cap(s), s)
}

// PrintMap 打印 map (调试用)
func PrintMap(name string, m map[string]int) {
	fmt.Printf("%s: %v\n", name, m)
}

// BenchmarkSort 基准测试模板
func BenchmarkSort(b *testing.B) {
	// TODO: 性能测试
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// RateLimiter 示例
	limiter := NewRateLimiter(10, 20)
	for i := 0; i < 25; i++ {
		if limiter.Allow() {
			fmt.Println("Request allowed")
		} else {
			fmt.Println("Request denied")
		}
	}

	// SafeMap 示例
	sm := NewSafeMap()
	sm.Store("key1", "value1")
	if val, ok := sm.Load("key1"); ok {
		fmt.Println("Found:", val)
	}

	// TaskExecutor 示例
	executor := NewTaskExecutor(context.Background(), 100)
	executor.Start(5)
	for i := 0; i < 10; i++ {
		executor.Submit(func() {
			fmt.Println("Task executed")
		})
	}
	executor.Stop()
}
