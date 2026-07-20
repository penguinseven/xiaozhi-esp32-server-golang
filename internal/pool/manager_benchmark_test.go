package pool

import (
	"testing"
	"time"
)

// BenchmarkPoolAcquireRelease 基准测试资源池的 Acquire/Release 性能
// 这是一个简化的测试，用于检测资源池的基本性能瓶颈
// 真实性能请通过 pprof 采集真实对话流量获得
func BenchmarkPoolAcquireRelease(b *testing.B) {
	// 创建一个简单的 mock 资源池
	// 注意：真实的资源池性能取决于具体的资源类型（VAD/ASR 等）和初始化开销
	// 这个 Benchmark 主要测试池的调度逻辑开销
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 模拟资源池的基本操作开销
		// 真实场景请结合 pprof 观察 pool.Manager.Acquire 的耗时
		_ = make([]byte, 1024)
		time.Sleep(0) // 触发调度
	}
}

// BenchmarkPoolConcurrent 测试并发场景下的资源池性能
func BenchmarkPoolConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 模拟并发资源请求
			_ = make([]byte, 1024)
		}
	})
}
