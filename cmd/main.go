package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Go 工程师面试准备练习")
	fmt.Println("========================================")
	fmt.Println()

	// 运行面试题示例
	runExamples()
}

// RateLimiter 令牌桶限流器
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

func runExamples() {
	fmt.Println("【示例】令牌桶限流器 (Rate Limiter)")
	fmt.Println("----------------------------------------")

	limiter := NewRateLimiter(10, 20) // 每秒10个令牌，容量20
	fmt.Printf("限流器配置: rate=%d req/s, capacity=%d\n\n", 10, 20)

	fmt.Println("发送 25 个请求:")
	successCount := 0
	for i := 0; i < 25; i++ {
		if limiter.Allow() {
			fmt.Printf("  [%2d] ✓ 允许\n", i+1)
			successCount++
		} else {
			fmt.Printf("  [%2d] ✗ 拒绝\n", i+1)
		}
	}
	fmt.Printf("\n成功: %d, 拒绝: %d\n", successCount, 25-successCount)

	fmt.Println("\n========================================")
	fmt.Println("更多练习请查看各目录下的测试文件:")
	fmt.Println("  - basics/      Go 语言基础练习")
	fmt.Println("  - concurrency/ 并发编程练习")
	fmt.Println("  - memory/      内存管理练习")
	fmt.Println("  - advanced/    高级特性练习")
	fmt.Println("  - interview/   面试高频题")
	fmt.Println("  - algorithm/   算法练习")
	fmt.Println("========================================")
	fmt.Println("\n运行测试命令:")
	fmt.Println("  go test -v ./basics/         # 基础练习")
	fmt.Println("  go test -v ./concurrency/    # 并发练习")
	fmt.Println("  go test -v ./interview/      # 面试题练习")
	fmt.Println("\n查看详细学习计划:")
	fmt.Println("  cat README.md")
	fmt.Println("========================================")
}
