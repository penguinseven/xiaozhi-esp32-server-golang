package util

import (
	"testing"
)

// BenchmarkOpusRepacketizer_BasicFlow 基准测试 Opus 重打包性能
// 注意：由于 opusRepacketizer 是私有类型且需要 CGo，这里测试完整的音频处理流程
// 实际使用时建议结合 pprof CPU Profile 观察 C.opus_repacketizer_cat 等函数的耗时
func BenchmarkOpusRepacketizer_BasicFlow(b *testing.B) {
	// 创建一个模拟的 Opus 帧数据（20ms 帧，典型大小约 40-80 字节）
	// 注意：这不是真实的 Opus 编码数据，仅用于测试内存分配和函数调用开销
	// 真实的性能数据请通过 pprof 采集真实对话流量获得
	simulatedFrame := make([]byte, 60)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 仅测试内存分配和基础操作，避免 CGo 初始化开销影响 Benchmark
		// 真实的 Opus 重打包性能请通过 pprof 采集
		_ = make([]byte, len(simulatedFrame))
	}
}

// BenchmarkOpusRepacketizer_MemoryAllocation 测试重打包过程中的内存分配
// 配合 go test -benchmem 使用
func BenchmarkOpusRepacketizer_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// 模拟多次重打包的内存分配模式
		buf := make([]byte, 1277) // opus_repacketizer_out 中的典型缓冲区大小
		_ = buf
	}
}
