package main

import (
	"fmt"
	"interview/concurrency"
	"interview/interview"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Go 工程师面试准备练习")
	fmt.Println("========================================")
	fmt.Println()

	// 运行面试题示例
	runExamples()
}

func runExamples() {
	fmt.Println("【示例1】限流器 (Rate Limiter)")
	fmt.Println("----------------------------------------")

	limiter := interview.NewRateLimiter(10, 20) // 每秒10个令牌，容量20
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
	fmt.Printf("\n成功: %d, 拒绝: %d\n\n", successCount, 25-successCount)

	// 等待令牌恢复
	fmt.Println("等待 1 秒后令牌恢复...")
	// 实际场景中可以使用 time.Sleep

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
	fmt.Println()

	// 导入并发包（避免 unused 错误）
	_ = concurrency.TestWorkerPool
}
