#!/bin/bash

# 运行所有 Benchmark 测试脚本

echo "=========================================="
echo "  运行所有 Benchmark 测试"
echo "=========================================="
echo ""

OUTPUT_FILE="./benchmark_results.txt"

echo "输出文件: $OUTPUT_FILE"
echo ""

{
    echo "=========================================="
    echo "  Benchmark 测试结果"
    echo "  生成时间: $(date)"
    echo "=========================================="
    echo ""

    echo "--- 1. 句子切分性能 ---"
    go test -bench="BenchmarkExtractSmartSentences" -benchmem -count=3 ./internal/util/... 2>&1 | grep -E "^Benchmark|^ok"
    echo ""

    echo "--- 2. Opus 重打包性能 ---"
    go test -bench="BenchmarkOpusRepacketizer" -benchmem -count=3 ./internal/util/... 2>&1 | grep -E "^Benchmark|^ok"
    echo ""

    echo "--- 3. 资源池性能 ---"
    go test -bench="BenchmarkPool" -benchmem -count=3 ./internal/pool/... 2>&1 | grep -E "^Benchmark|^ok"
    echo ""

    echo "--- 4. 现有其他 Benchmark ---"
    go test -bench="Benchmark" -benchmem -count=1 ./... 2>&1 | grep -E "^Benchmark" | grep -v "^BenchmarkExtractSmartSentences\|^BenchmarkOpusRepacketizer\|^BenchmarkPool"
    echo ""

    echo "=========================================="
    echo "  测试完成"
    echo "=========================================="
} | tee "$OUTPUT_FILE"

echo ""
echo "结果已保存到: $OUTPUT_FILE"

