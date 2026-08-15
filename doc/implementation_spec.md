# 企业级改造实现规格

> 基线：[wayfinder:map #1](https://github.com/penguinseven/xiaozhi-esp32-server-golang/issues/1) · 9 个决策工单（#2-#10）全部关闭 · 2026-08-13

## 前提

- 词汇以 `CONTEXT.md` 为准（Device Config / System Config / Config Source / Config Reader / Settings / AI Provider）
- 配置所有权已接受（ADR 0001）：类型化 `Settings` + 窄 `Config Reader` + `Config Source` 由 Init 构造一次
- 蓝图：`doc/engineering_standardization_plan.md`（五阶段，布局/配置/分层评分）
- 两个根因接缝均已决策：配置所有权（ADR 0001+#2+#3）、LLM eino 形状（#4）

## 实现阶段

### Phase 1 — 清场

**#5 死模块清理**（priority: 最高，无依赖，先做以缩小变更面）

| 动作 | 说明 |
|------|------|
| 删除 `internal/config/` | 零 import（主模块），`asr_server/` 独立 go module 不受影响 |
| 删除 `internal/domain/eventbus/` | 5 文件，零外部引用 |
| 删除 `internal/domain/memory/` | mem0/memobase/memos/nomemo/llm_memory，零外部引用 |
| 删除 `internal/domain/llm/test/` | 零引用 |
| 删除 `eino_llm/example.go` | 510 行示例代码在 production 包 |
| 删除 `streamtransform.Registry` | 零 Register 调用，保留 Pipeline/Item/Context/Kind |
| 删除 `llm.go` 的 `ConvertMCPToolsToEinoTools` | 零调用（#4 覆盖） |
| 清理 `go.mod` | 移除不再需要的依赖 |

**验收**：`go build ./...` 通过，无 broken import。

---

### Phase 2 — 配置基础

**#2 配置解析链** + **#3 Config Source 生命周期**（并行推进，同一模块）

**#2 配置解析链**
| 动作 | 说明 |
|------|------|
| 实现分键类优先级 | 普通键 `defaults→base→local→env→remote`；敏感键 `local→remote→env` |
| 逐键 `BindEnv` 白名单 | `XIAOZHI_<KEY>`，初始=密钥清单 |
| local 改真 merge | 不再覆盖，深度合并 |
| `Resolve()` 顺序 | defaults→base→local→env→remote→热更段 |
| 移除 3 处明文口令 | `config/mqtt_config.json`、`config/config.yaml`、`build/common/manager.json` |
| 统一 `config_provider.type` 默认 | 消除 `config.yaml`(memory) 与 `config_init.go`(redis) 不一致 |
| 敏感键 env 来源安全约束 | 仅允许 `XIAOZHI_` 前缀 + 显式白名单 |

**#3 Config Source 生命周期**
| 动作 | 说明 |
|------|------|
| Config Source 由 Init 构造一次 | 单例持有（`GetProvider` 返回单例），不再每次现造 HTTP client |
| 宽接口拆 4 个窄接口 | `ActivationReader` / `DeviceConfigReader` / `SystemConfigReader` / `ConfigEventSink` |
| 11 处调用点迁移 | 消费方改用对应窄接口 |
| `provider_test.go` 重写 | 删陈旧测试，按窄接口重写 |
| provider type 收敛 | 常量化到 `internal/constants` |

**验收**：`go test ./internal/pkg/config/...` 通过，消费方编译通过。

---

### Phase 3 — 核心接缝

**#4 LLM 接缝** + **#7 资源池泛型** + **#6 CGo 隔离**（并行推进，写集不重叠）

**#4 LLM 接缝领域化**（最大变更面）
| 动作 | 说明 |
|------|------|
| 定义 `llm.Message`/`llm.Tool`/`llm.ToolCall`/`llm.LLMResponse` | `internal/domain/llm/types.go` |
| `LLMProvider` 收窄 | `ResponseWithContext`（→`llm.*`）+ `ResponseWithVllm`（保留）+ `Close`/`IsValid` |
| 删除 `GetModelInfo`/`LLMFactory` | 零调用 |
| eino 双向映射 | `eino_llm` 内 domain↔eino 映射（`Message`↔`schema.Message` 等） |
| 消费端迁移 | chat/llm.go、hooks、streamtransform、eventbus、mcp、memory → `llm.*` |
| eino import 约束 | 只允许 `eino_llm` adapter 内出现 |
| thinking transport | 留在 `eino_llm`，domain 只暴露 `HasReasoningContent` 窄能力 |

**#7 资源池泛型**
| 动作 | 说明 |
|------|------|
| `ResourceTypeOption[T]` 泛型化 | `WithCloseFunc(func(T) error)` 等 |
| 4 处注册迁移 | 12 段断言样板替换为无断言调用 |
| `Lifecycle[T]` 聚合 | `Close(T) error` + `IsValid(T) bool` + `Reset(T) error` |

**#6 CGo 隔离**
| 动作 | 说明 |
|------|------|
| 新建 `internal/domain/audio/codec` | 暴露纯 Go `FrameConverter` interface |
| CGo 实现 `//go:build cgo` | opus 编解码仅在有 libopus 编译时激活 |
| 纯 Go 回退 `//go:build !cgo` | PCM passthrough，CI 不崩 |
| `ten_vad_cgo.go` | `//go:build cgo && ten_vad` 双 tag，无 cgo 时跳过注册 |
| util 音频函数迁移 | `WavToOpus`/`NormalizeOpusSampleRate`/`opus_repacketizer` → `audio/codec` |
| `audio_handler.go` | 以 `FrameConverter` 取代 |

**验收**：`go build ./...` + `go test ./...` 通过（含 cgo 与 !cgo 两种 build tag）。

---

### Phase 4 — 收尾

**#9 TTS 帧管线**（依赖 #6）+ **#8 RAG runner** + **#10 热更编排**（依赖 #2/#3）

**#9 TTS 帧管线**
| 动作 | 说明 |
|------|------|
| `internal/domain/tts/frames/stream.go` | `StreamToFrames(stream, format, frameDuration)` 深模块 |
| 10+ provider 替换 | `CreateAudioDecoder+WithFormat+decoder.Run` → `StreamToFrames` 单行 |
| 格式矩阵集中 | mp3/wav/pcm/opus/ogg_opus 归一在一处 |
| 单测矩阵 | 格式组合 × 采样率 × 帧长 |

**#8 RAG runner**
| 动作 | 说明 |
|------|------|
| `internal/domain/rag/runner.go` | `searchKnowledgeBases(ctx, kbs, topK, searchOne)` 共享 runner |
| 三 searcher 瘦身 | 只保留 `searchOne`（HTTP body 构造） |
| 单测集中 | `runner_test.go` 覆盖并发/超时/部分失败 |

**#10 热更编排**
| 动作 | 说明 |
|------|------|
| `internal/app/server/configreload/` | `HotReloadManager` + `AppReloadCallbacks` 接口 |
| `cmd/server/main.go` 瘦身 | 只注册回调，diff→merge→reload 进模块 |
| `cmd/mqtt` 暂不复用 | 无热更需求，后续确需时接入 |

**验收**：全量 `go build ./...` + `go test ./...` 通过，`cmd/server` 启动正常。

---

## 执行顺序

```
Phase 1 (#5) ──────────────────────────────────────────────┐
                                                           │
Phase 2 (#2 + #3) ─────────────────────────────────────────┤
                                                           │
Phase 3 (#4 + #7 + #6) ────────────────────────────────────┤
│                                                          │
Phase 4 (#9←#6 + #8 + #10←#2/#3) ──────────────────────────┘
```

## 关联

- 蓝图回写：`doc/engineering_standardization_plan.md` 待补两个根因接缝条目
- 雾区待毕业：会话拆分、TurnCoordinator、ConversationStore、IConn 类型化、logger 深化、目录收敛
