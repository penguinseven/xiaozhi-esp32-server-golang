#!/bin/bash

# 阶段 0：性能基准测试采集脚本
# 使用方法：
# 1. 先启动服务并确保 pprof 已启用（config.yaml 中 server.pprof.enable: true）
# 2. 运行设备对话 5 分钟以上
# 3. 执行此脚本采集 profile

PPROF_ADDR="http://localhost:6060/debug/pprof"
OUTPUT_DIR="./profiles"
DURATION_SECONDS=300  # 5 分钟

mkdir -p "$OUTPUT_DIR"

echo "=========================================="
echo "  阶段 0：性能基准测试采集脚本"
echo "=========================================="
echo ""
echo "采集目标: $PPROF_ADDR"
echo "输出目录: $OUTPUT_DIR"
echo "采集时长: $DURATION_SECONDS 秒"
echo ""

# 1. CPU Profile
echo "[1/5] 采集 CPU Profile..."
go tool pprof -seconds=$DURATION_SECONDS -output="$OUTPUT_DIR/cpu_profile.pprof" "$PPROF_ADDR/profile"

# 生成 CPU 火焰图 SVG
echo "生成 CPU 火焰图..."
go tool pprof -svg "$OUTPUT_DIR/cpu_profile.pprof" > "$OUTPUT_DIR/cpu_profile.svg"

# 2. 内存 Profile (heap)
echo "[2/5] 采集 内存 Profile (heap)..."
go tool pprof -output="$OUTPUT_DIR/heap_profile.pprof" "$PPROF_ADDR/heap"
echo "生成 内存 火焰图..."
go tool pprof -svg "$OUTPUT_DIR/heap_profile.pprof" > "$OUTPUT_DIR/heap_profile.svg"

# 3. Goroutine Profile
echo "[3/5] 采集 Goroutine Profile..."
go tool pprof -output="$OUTPUT_DIR/goroutine_profile.pprof" "$PPROF_ADDR/goroutine"
echo "生成 Goroutine 火焰图..."
go tool pprof -svg "$OUTPUT_DIR/goroutine_profile.pprof" > "$OUTPUT_DIR/goroutine_profile.svg"

# 4. Block Profile (阻塞分析)
echo "[4/5] 采集 Block Profile (阻塞分析)..."
go tool pprof -output="$OUTPUT_DIR/block_profile.pprof" "$PPROF_ADDR/block" 2>/dev/null || echo "注意: Block Profile 可能需要在代码中启用 runtime.SetBlockProfileRate"

# 5. Mutex Profile (锁竞争)
echo "[5/5] 采集 Mutex Profile (锁竞争)..."
go tool pprof -output="$OUTPUT_DIR/mutex_profile.pprof" "$PPROF_ADDR/mutex" 2>/dev/null || echo "注意: Mutex Profile 可能需要在代码中启用 runtime.SetMutexProfileFraction"

echo ""
echo "=========================================="
echo "  采集完成！"
echo "=========================================="
echo ""
echo "输出文件列表:"
ls -lh "$OUTPUT_DIR/"
echo ""
echo "下一步操作:"
echo "1. 查看 CPU 热点: go tool pprof -top $OUTPUT_DIR/cpu_profile.pprof"
echo "2. 查看内存热点: go tool pprof -top $OUTPUT_DIR/heap_profile.pprof"
echo "3. 查看火焰图: 打开 $OUTPUT_DIR/cpu_profile.svg"
echo ""
echo "瓶颈判定标准:"
echo "- CGo 调用占比 > 15%: 高优先级，考虑阶段 4 解耦"
echo "- Channel 阻塞 > 10%: 中优先级，考虑并发优化"
echo "- 资源池 Acquire 等待 > 5%: 中优先级，考虑池大小优化"
echo "- GC 压力（音频帧分配）> 10%: 中优先级，考虑内存池"

