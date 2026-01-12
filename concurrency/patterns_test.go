package concurrency

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// 练习1：Worker Pool 模式
// 经典的并发模式，用于限制并发数量

func TestWorkerPool(t *testing.T) {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// 启动 3 个 worker
	numWorkers := 3
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range jobs {
				// 模拟处理任务
				time.Sleep(100 * time.Millisecond)
				results <- j * 2
				t.Logf("Worker %d processed job %d", id, j)
			}
		}(w)
	}

	// 发送任务
	go func() {
		for j := 1; j <= 9; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	// 等待所有 worker 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	for r := range results {
		t.Logf("Result: %d", r)
	}
}

// 练习2：Fan-Out / Fan-In 模式
// Fan-Out: 将输入分发到多个 goroutine
// Fan-In: 将多个 goroutine 的结果聚合

func TestFanOutFanIn(t *testing.T) {
	// TODO: 实现 Fan-Out/Fan-In 模式
	// 场景：并行处理多个任务，然后汇总结果
}

// 练习3：Context 超时控制
func TestContextTimeout(t *testing.T) {
	// 3.1 WithTimeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Log("Operation completed")
	case <-ctx.Done():
		t.Logf("Operation timed out: %v", ctx.Err())
	}

	// TODO: 实现 WithDeadline
	// TODO: 实现 WithCancel
}

// 练习4：Or-Done Channel 模式
// 多个 channel 中任意一个完成即返回

func TestOrDone(t *testing.T) {
	or := func(channels ...<-chan interface{}) <-chan interface{} {
		// TODO: 实现 or-done 模式
		// 当任意一个 channel 关闭时，返回的 channel 也关闭
		return nil
	}

	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Second),
		sig(5*time.Second),
		sig(1*time.Second),
		sig(3*time.Second),
	)
	t.Logf("Done after %v", time.Since(start))
}

// 练习5：Rate Limiting (令牌桶)
func TestRateLimiter(t *testing.T) {
	// TODO: 实现一个令牌桶限流器
	// 要求：
	// - 每秒生成 N 个令牌
	// - 桶容量最大 M 个令牌
	// - 请求时消耗令牌，无令牌时等待或拒绝
}

// 练习6：Pipeline 模式
func TestPipeline(t *testing.T) {
	// Generator: 生成数据
	generator := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				out <- n
			}
		}()
		return out
	}

	// Square: 计算平方
	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * n
			}
		}()
		return out
	}

	// 使用 pipeline
	for n := range square(generator(1, 2, 3, 4, 5)) {
		t.Logf("Pipeline result: %d", n)
	}
}

// 练习7：检测 goroutine 泄漏
func TestGoroutineLeak(t *testing.T) {
	// 错误示例：goroutine 永远阻塞（仅作演示，不调用）
	_ = func() {
		ch := make(chan int)
		go func() {
			val := <-ch // 如果没有发送者，永远阻塞
			t.Logf("Received: %d", val)
		}()
		// ch 没有发送者，goroutine 泄漏
	}

	// 正确做法：使用 context 取消或 buffered channel
	correctFunc := func() {
		ch := make(chan int, 1) // buffered channel
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			select {
			case val := <-ch:
				t.Logf("Received: %d", val)
			case <-ctx.Done():
				t.Log("Goroutine cancelled")
				return
			}
		}()

		cancel() // 取消 goroutine
		time.Sleep(100 * time.Millisecond)
	}

	correctFunc()
}

// 练习8：Mutex vs RWMutex
func TestMutexTypes(t *testing.T) {
	// 8.1 普通 Mutex
	var mu sync.Mutex
	var count int

	_ = func() { // increment 示例，仅作演示
		mu.Lock()
		defer mu.Unlock()
		count++
	}

	// 8.2 RWMutex: 读多写少场景性能更好
	var rwmu sync.RWMutex
	var data map[string]string

	_ = func(key string) string { // read 示例，仅作演示
		rwmu.RLock()
		defer rwmu.RUnlock()
		return data[key]
	}

	_ = func(key, value string) { // write 示例，仅作演示
		rwmu.Lock()
		defer rwmu.Unlock()
		if data == nil {
			data = make(map[string]string)
		}
		data[key] = value
	}

	// TODO: 读写锁的使用场景
	// - 适合读多写少的场景
	// - 多个读者可以同时持有读锁
	// - 写者是排他的
}

// 练习9：Once 单例模式
func TestOnceSingleton(t *testing.T) {
	type Singleton struct {
		Data string
	}

	var instance *Singleton
	var once sync.Once

	getInstance := func() *Singleton {
		once.Do(func() {
			t.Log("Initializing singleton...")
			instance = &Singleton{Data: "singleton"}
		})
		return instance
	}

	// 多次调用只初始化一次
	getInstance()
	getInstance()
	getInstance()

	t.Logf("Singleton: %+v", instance)
}

// 练习10：Pool 对象池
func TestPool(t *testing.T) {
	// sync.Pool 用于临时对象的复用，减少 GC 压力
	pool := &sync.Pool{
		New: func() interface{} {
			t.Log("Creating new object")
			return make([]byte, 1024)
		},
	}

	// 获取对象
	buf1 := pool.Get().([]byte)
	t.Logf("Got buf1: len=%d", len(buf1))

	// 归还对象
	pool.Put(buf1)

	// 再次获取，可能复用之前归还的
	buf2 := pool.Get().([]byte)
	t.Logf("Got buf2: len=%d", len(buf2))

	// TODO: Pool 的注意事项
	// - Pool 中的对象可能被 GC 自动清除
	// - 不要在 Pool 中存敏感信息
	// - Put 前要重置对象状态
}

// 练习11：Cond 条件变量
func TestCond(t *testing.T) {
	// 条件变量用于协调多个 goroutine 的执行顺序
	var (
		ready bool
		mu    sync.Mutex
		cond  = sync.NewCond(&mu)
	)

	// 等待者
	go func() {
		mu.Lock()
		for !ready {
			cond.Wait() // 等待通知
		}
		t.Log("Goroutine is ready!")
		mu.Unlock()
	}()

	// 通知者
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	ready = true
	cond.Signal() // 唤醒一个等待者
	mu.Unlock()

	time.Sleep(100 * time.Millisecond)
}

// 练习12：Select 的随机性
func TestSelectRandomness(t *testing.T) {
	ch1 := make(chan int, 10)
	ch2 := make(chan int, 10)

	// 两个 channel 都就绪
	ch1 <- 1
	ch2 <- 2

	counts := make(map[int]int)

	// 多次运行，观察 select 的随机性
	for i := 0; i < 1000; i++ {
		select {
		case <-ch1:
			counts[1]++
		case <-ch2:
			counts[2]++
		}
		// 把数据放回去
		ch1 <- 1
		ch2 <- 2
	}

	fmt.Printf("Select counts: %v\n", counts)
	// TODO: 理解 select 当多个 case 就绪时的随机选择
}

// 练习13：Channel 关闭与检测
func TestChannelClose(t *testing.T) {
	ch := make(chan int, 3)

	go func() {
		ch <- 1
		ch <- 2
		close(ch) // 关闭 channel
	}()

	// 使用 range 自动检测关闭
	for v := range ch {
		t.Logf("Received: %d", v)
	}

	// 使用 v, ok := <-ch 检测
	v, ok := <-ch
	t.Logf("After close: v=%d, ok=%v", v, ok)

	// TODO: 向已关闭的 channel 发送会 panic？
	// 从已关闭的 channel 读取会返回零值和 false
}

// 练习14：Buffered vs Unbuffered Channel
func TestChannelBuffer(t *testing.T) {
	// 无缓冲 channel：同步通信，发送者会阻塞直到有接收者
	unbuffered := make(chan int)
	go func() {
		unbuffered <- 1 // 如果没有接收者，这里会永久阻塞
	}()
	val := <-unbuffered
	t.Logf("Unbuffered received: %d", val)

	// 有缓冲 channel：异步通信，缓冲区满时发送者才阻塞
	buffered := make(chan int, 2)
	buffered <- 1
	buffered <- 2
	// buffered <- 3 // 会阻塞，因为缓冲区已满
	t.Logf("Buffered len: %d, cap: %d", len(buffered), cap(buffered))
}
