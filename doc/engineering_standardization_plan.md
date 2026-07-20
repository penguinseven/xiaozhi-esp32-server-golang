# 项目工程化标准与重构方案

本指南对标 **Go 官方工程规范**、**12-Factor App**、**Clean / Hexagonal Architecture** 及 **golang-standards/project-layout**，对 `xiaozhi-esp32-server-golang` 项目的代码布局、配置规划、功能划分进行深度解剖，评估其合理性、标准性，并给出具体的重构设计蓝图。

所有评估结论均基于对实际代码库的扫描验证（Go 1.24.2、三个独立 Go Module、CGo 依赖、viper 配置机制、eino 框架引入范围等）。

---

## 一、现状评估与痛点分析

### 1. 项目布局（Layout）

* **合理性：7.5 / 10** -- 功能模块化清晰，核心领域划分明确，具备企业级项目的基本骨架。
* **标准性：5.5 / 10** -- 存在顶级目录污染、Monorepo 缺乏协同、内部目录语义混杂等问题。

**核心痛点：**

1. **顶级目录污染**
   `logger/`、`lib/`、`constants/` 被直接置于项目根目录。按 golang-standards 规范，根目录应只保留 `cmd/`、`internal/`、`config/`、`go.mod`、`Makefile` 等必要文件。

2. **Monorepo 缺乏协同**
   项目包含三个独立 Go Module（根目录主服务 `xiaozhi-esp32-server-golang`、`manager/backend` 模块 `xiaozhi/manager/backend`、Git 子模块 `asr_server` 模块 `voice_server`），但：
   - **未配置 `go.work`**（Go 1.20+ 工作区），全靠 `replace` 指令串联：
     - 主模块 `go.mod`：`replace xiaozhi/manager/backend => ./manager/backend`、`replace voice_server => ./asr_server`
     - `manager/backend/go.mod`：`replace xiaozhi-esp32-server-golang => ../..`（双向 replace）
   - IDE 无法自动跨 Module 跳转，本地联合调试需手动改 `replace` 路径。

3. **内部目录语义混杂**
   `internal/pkg/hooks/`（抽象插件总线）与 `internal/util/`（业务工具集）命名语义模糊，虽实际职责不重叠，但 `pkg` vs `util` 边界不清晰。`internal/util/` 是工具函数大杂烩（音频处理、加密、队列、句子切分、资源池混在一起），缺乏按主题分包。

### 2. 配置规划（Configuration）

* **合理性：6.5 / 10** -- 支持多配置 Provider 模式（memory / manager / redis），动态配置合并机制设计良好。
* **标准性：5.0 / 10** -- 本地配置覆盖逻辑反直觉、配置格式碎片化、缺乏环境变量支持。

**核心痛点：**

1. **本地配置覆盖逻辑反直觉**
   `config.local.yaml` 是对 `config.yaml` 的**完整替代**，而非**合并覆盖（Merge/Override）**。由 Makefile 决定二选一：
   ```make
   LOCAL_CONFIG ?= config/config.local.yaml
   CONFIG       ?= $(if $(wildcard $(LOCAL_CONFIG)),$(LOCAL_CONFIG),config/config.yaml)
   ```
   - 默认配置新增项时，本地配置因缺项会报运行错误。
   - 对比：远程配置（manager/redis）已使用 `viper.MergeConfigMap()` 做真正的合并，本地却做不到。

2. **配置格式与源碎片化**
   项目并存多种配置格式：
   - `config/config.yaml`（主服务，viper + YAML）
   - `config/mqtt_config.json`（MQTT 工具，viper + JSON）
   - `asr_server/config.json`（ASR 服务，viper + JSON，支持 fsnotify 热重载）
   - `manager/backend/config.json`（管理后台，非 viper，原生 JSON）
   - `internal/config/config.go`（遗留 JSON Config，疑似未在主流程使用）

3. **违背 12-Factor 环境变量原则**
   项目不支持通过 OS 环境变量注入敏感信息（如 `DEEPSEEK_API_KEY`、`MYSQL_PASSWORD`）。对 Docker/K8s 容器化部署极不友好，也容易导致 API Key 被误提交（已在本次开发中触发 GitHub Push Protection 拦截）。

### 3. 功能划分与分层（Functional Partitioning）

* **合理性：7.0 / 10** -- Provider 抽象层设计良好，方便对接多家 ASR/TTS 服务。
* **标准性：5.5 / 10** -- 框架实现泄露到领域层，CGo 强耦合导致编译笨重。

**核心痛点：**

1. **具体框架向 Domain 泄露**
   `cloudwego/eino` 框架被广泛引入 `internal/domain/` 下多个子目录：

   | 子目录 | 是否依赖 eino | 泄露程度 |
   |--------|-------------|---------|
   | `llm/`（base, common, eino_llm, coze_llm, dify_llm） | **是（重度）** | eino/schema + eino/components/model + eino-ext |
   | `mcp/`（local_manager, global_manage, mcp_tool 等） | **是（重度）** | eino/components/tool + eino/schema |
   | `memory/`（base, llm_memory, mem0, memobase, memos, nomemo） | **是** | eino/schema |
   | `chat/hooks/`、`chat/streamtransform/` | **是** | eino/schema |
   | `eventbus/` | **是** | eino/schema |
   | `config/manager/` | **是** | eino/schema + eino/components/tool |
   | `asr/`、`tts/`、`vad/`、`speaker/`、`audio/` | 否 | 纯原生协议 |

   按 Clean Architecture，领域层应保持 100% 纯粹，不依赖任何第三方业务框架。Eino 是具体的 LLM 编排框架，属于**外层基础设施/适配器**（Infrastructure/Adapter）。

2. **CGo 编译屏障严重**
   CGo 集中在两处：
   - **Opus 音频编解码**：`internal/util/opus_repacketizer.go`（`#cgo pkg-config: opus`），Makefile 强制 `CGO_ENABLED=1`
   - **TEN-VAD 语音活动检测**：`internal/domain/vad/ten_vad/ten_vad_cgo.go`（`//go:build cgo`，链接 `lib/ten-vad/` 下的 `.so`/`.dll`/`.framework`）

   主服务因 Opus 编解码必须开 CGo，导致整个主服务的交叉编译非常笨重，丧失了 Go 跨平台快速分发的优势。

3. **双套日志体系不一致**
   - 主服务 + manager backend：`logrus` + `nested-logrus-formatter` + `file-rotatelogs`
   - `asr_server` 子模块：Go 1.21+ 标准库 `log/slog` + `lumberjack` 轮转
   - 两套日志格式不统一，跨模块排查问题时需适配两种日志格式。
   - 无 OpenTelemetry 分布式追踪。

---

## 二、重构方案

### 1. 布局重构：遵循 golang-standards 规范

**参照标准：** [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

**重构方案：**

- **收拢顶级目录**：
  - `logger/` -> `internal/pkg/logger/`
  - `constants/` -> `internal/constants/`
  - `lib/` -> `build/lib/`（平台相关的 `.so`/`.dll`/`.framework` 属于构建产物）

- **启用 Go Workspaces**：
  在根目录创建 `go.work`，统一管理三个独立 Go Module：
  ```
  go 1.24

  use (
      .
      ./manager/backend
      ./asr_server
  )
  ```
  替代现有 3 处 `replace` 指令，IDE 自动跨 Module 跳转。

- **规范 `cmd/` 的职责**：
  `cmd/server/main.go` 只做：命令行参数解析、加载配置、优雅退出信号监听。复杂的初始化逻辑全部内聚到 `internal/app/`。

- **拆分 `internal/util/`**：
  按主题分包，消除工具函数大杂烩：
  - `internal/pkg/audio/` -- 音频处理（opus、ogg、重打包）
  - `internal/pkg/crypto/` -- 加密、密码签名
  - `internal/pkg/queue/` -- 队列、资源池
  - `internal/pkg/text/` -- 句子切分、文本处理

### 2. 配置重构：层次化合并与 12-Factor 标准

**参照标准：** [The Twelve-Factor App - III. Config](https://12factor.net/config)

**重构方案：**

- **多层级叠加覆盖（Merge）**：
  利用 Viper 已有的 `MergeConfigMap` 能力，改为分层加载：
  ```
  config/config.yaml (默认，Git 跟踪)
    ↑ 合并
  config/config.local.yaml (本地差异，GitIgnored)
    ↑ 合并
  OS 环境变量 (DEEPSEEK_API_KEY 等)
    ↑ 合并
  数据库 / Redis 远程配置 (manager / redis provider)
  ```
  本地只需写差异项，无需完整复制。

- **绑定环境变量**：
  ```go
  viper.SetEnvPrefix("XIAOZHI")
  viper.AutomaticEnv()
  viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
  ```
  ```bash
  export XIAOZHI_LLM_DEEPSEEK_API_KEY="sk-xxxx"
  export XIAOZHI_REDIS_PASSWORD="xxx"
  ```
  无缝支持 Docker/K8s 部署，从根本上避免 API Key 被提交到代码库。

- **统一配置格式**：
  全部使用 YAML（包括 `asr_server` 和 `manager/backend`），消除 JSON/YAML 并存问题。

- **清理遗留配置**：
  移除 `internal/config/config.go`（遗留 JSON Config，未在主流程使用）。

### 3. 架构重构：六边形架构（Ports and Adapters）

**参照标准：** [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)

**重构方案：**

- **净化 Domain（领域层）**：
  `internal/domain/` 下只定义**纯粹的 Interface（Ports/端口）**和**领域核心模型**：
  ```go
  // internal/domain/llm/provider.go
  type LLMProvider interface {
      Chat(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (<-chan *schema.Message, error)
  }
  ```
  此文件不得 import 任何 `eino` 包。

- **下放实现到 Infrastructure（适配器层）**：
  | 当前位置 | 目标位置 | 说明 |
  |---------|---------|------|
  | `internal/domain/llm/eino_llm/` | `internal/infrastructure/llm/eino/` | Eino 框架实现 |
  | `internal/domain/llm/coze_llm/` | `internal/infrastructure/llm/coze/` | Coze 实现 |
  | `internal/domain/llm/dify_llm/` | `internal/infrastructure/llm/dify/` | Dify 实现 |
  | `internal/domain/vad/ten_vad/` | `internal/infrastructure/vad/ten/` | TEN-VAD CGo 实现 |
  | `internal/domain/vad/silero_vad/` | `internal/infrastructure/vad/silero/` | Silero 实现 |
  | `internal/domain/mcp/` | `internal/infrastructure/mcp/eino/` | Eino MCP 实现 |

- **渐进式迁移策略**：
  不一次性重构，按模块逐步迁移。每次迁移一个 Provider，确保测试通过再继续。优先迁移 `llm/eino_llm/`（泄露最严重），再迁移 `mcp/`、`memory/`。

### 4. CGo 解耦：进程间通信替代

**当前状态：**
- Opus 编解码必须 CGo -> 主服务无法纯 Go 编译
- TEN-VAD 已用 `//go:build cgo` 隔离，但仍在主服务目录下

**重构方案：**

- **Opus 编解码**：替换为纯 Go 实现 `github.com/hraban/opus` 已在用，但 `opus_repacketizer.go` 仍直接 `#cgo`。改为通过 `hraban/opus` 库间接调用，或用纯 Go 的 [opus-go](https://github.com/hraban/opus) 重打包接口封装，隔离 CGo 到单一文件。

- **TEN-VAD**：已通过 `//go:build cgo` 隔离。进一步将 CGo 代码移到 `asr_server` 子模块，主服务通过 gRPC/Unix Socket 调用，实现 100% 纯 Go 编译。

- **asr_server 独立部署**：主服务通过 gRPC / Local IPC Unix Socket / WebSocket 与 `asr_server` 交互，主服务获得纯 Go 极速编译体验。

### 5. 日志与可观测性统一

**重构方案：**

- **统一日志框架**：全项目迁移到 Go 1.21+ 标准库 `log/slog`（`asr_server` 已在用），替换主服务的 logrus。
- **结构化日志格式**：统一 JSON Handler，便于 ELK/Loki 采集。
- **引入 OpenTelemetry**：添加分布式追踪，关键链路（hello -> ASR -> LLM -> TTS）加 span，便于延迟分析。

---

## 三、重构后项目布局蓝图

```
xiaozhi-esp32-server-golang/
├── go.work                         # Go Workspaces（多 Module 联合调试）
├── go.mod                          # 主服务 Module (xiaozhi-esp32-server-golang)
├── go.sum
├── Makefile
├── README.md
├── .golangci.yml                   # 代码质量检查配置
│
├── cmd/                            # 执行入口
│   ├── server/                     # 主服务入口
│   │   └── main.go                 # 仅做：参数解析、配置加载、信号监听
│   ├── mqtt/                       # MQTT 工具入口
│   └── mock_ai_server/             # Mock AI 服务入口
│
├── config/                         # 默认配置文件
│   └── config.yaml                 # Git 跟踪，不含敏感信息
│
├── data/                           # 本地持久化（GitIgnored）
│   ├── sqlite/                     # SQLite 数据库
│   ├── storage/                    # 声纹向量、上传文件
│   └── logs/                       # 日志文件
│
├── internal/                       # 主服务核心（受 internal 保护）
│   ├── app/                        # 应用初始化
│   │   └── server/                 # 服务启动、端口绑定、优雅退出
│   │
│   ├── domain/                     # 【领域层】无外部框架依赖
│   │   ├── session/                # 会话状态机、信令协议
│   │   ├── asr/                    # type AsrProvider interface
│   │   ├── llm/                    # type LLMProvider interface
│   │   ├── tts/                    # type TtsProvider interface
│   │   ├── vad/                    # type VadProvider interface
│   │   ├── chat/                   # 会话编排（纯领域逻辑）
│   │   └── message/                # 消息模型定义
│   │
│   ├── infrastructure/             # 【适配器层】具体框架实现
│   │   ├── llm/
│   │   │   ├── eino/               # Eino 框架 LLM 实现
│   │   │   ├── coze/               # Coze LLM 实现
│   │   │   └── dify/               # Dify LLM 实现
│   │   ├── asr/
│   │   │   ├── doubao/             # 豆包 ASR
│   │   │   ├── funasr/             # FunASR
│   │   │   ├── aliyun/             # 阿里云 ASR
│   │   │   └── xunfei/             # 讯飞 ASR
│   │   ├── tts/
│   │   │   ├── edge/                # Edge TTS（免费）
│   │   │   ├── doubao/             # 豆包 TTS
│   │   │   ├── cosyvoice/          # CosyVoice
│   │   │   └── xunfei/             # 讯飞 TTS
│   │   ├── vad/
│   │   │   ├── ten/                # TEN-VAD (CGo)
│   │   │   ├── silero/             # Silero VAD
│   │   │   └── webrtc/             # WebRTC VAD
│   │   ├── mcp/
│   │   │   └── eino/               # Eino MCP 工具管理
│   │   ├── memory/
│   │   │   ├── redis/              # Redis 短期记忆
│   │   │   ├── memobase/           # Memobase 长期记忆
│   │   │   └── mem0/               # Mem0 长期记忆
│   │   └── db/                     # GORM / SQLite 存储适配
│   │
│   ├── pkg/                        # 通用公共库（全项目共享）
│   │   ├── logger/                 # slog 结构化日志
│   │   ├── config/                 # 层次化配置加载器
│   │   ├── audio/                  # 音频处理（opus、ogg）
│   │   ├── crypto/                 # 加密、签名
│   │   ├── queue/                  # 队列、资源池
│   │   ├── text/                   # 句子切分、文本处理
│   │   └── hooks/                  # 插件/Hook 总线
│   │
│   └── constants/                  # 常量定义
│
├── manager/                        # 管理后台（子 Module）
│   ├── backend/                    # Go API
│   └── frontend/                   # Vue 3 前端
│
├── asr_server/                     # ASR/声纹微服务（子 Module，独立进程）
│   └── main.go
│
├── build/                          # 构建产物
│   ├── lib/                        # 平台相关库（.so/.dll/.framework）
│   ├── common/                     # 打包用公共配置
│   └── models/                     # VAD/声纹 onnx 模型
│
├── docker/                         # Dockerfile
├── doc/                            # 文档
├── ai_doc/                         # 架构设计文档
└── test/                           # 集成测试与压测工具
```

---

## 四、改造收益

| 维度 | 当前状态 | 重构后 |
|------|---------|--------|
| **编译速度** | 主服务强制 CGo，交叉编译笨重 | 主服务 100% 纯 Go，秒级编译 |
| **跨 Module 调试** | 3 处 replace，IDE 跳转断裂 | go.work 统一管理，无缝跳转 |
| **配置管理** | 本地配置完整替代，容易缺项 | 层次化合并，本地只写差异 |
| **敏感信息安全** | API Key 写在 YAML，易误提交 | 环境变量注入，Push Protection 不再拦截 |
| **领域层纯净度** | eino 框架泄露到 6 个 domain 子目录 | 领域层零框架依赖，纯接口定义 |
| **单元测试** | 领域层耦合框架，Mock 困难 | 领域层纯接口，Mock 简单 |
| **日志一致性** | logrus + slog 双套并行 | 统一 slog + JSON 格式 |
| **可观测性** | 无分布式追踪 | OpenTelemetry 全链路追踪 |

---

## 五、迁移路线图

采用**渐进式迁移**，不中断现有功能：

```
阶段 1：基础设施 (低风险)
  ├── 创建 go.work，替代 replace 指令
  ├── 收拢顶级目录 (logger/, constants/, lib/)
  └── 统一日志到 slog

阶段 2：配置体系 (中风险)
  ├── 改 config.local.yaml 为合并模式
  ├── 添加环境变量绑定 (AutomaticEnv)
  └── 清理遗留配置代码

阶段 3：领域层净化 (高风险，逐模块迁移)
  ├── 迁移 llm/eino_llm/ -> infrastructure/llm/eino/
  ├── 迁移 mcp/ -> infrastructure/mcp/eino/
  ├── 迁移 memory/ -> infrastructure/memory/
  └── 每次迁移一个 Provider，确保测试通过

阶段 4：CGo 解耦 (高风险)
  ├── 隔离 Opus CGo 到单一文件
  ├── TEN-VAD 移到 asr_server
  └── 主服务通过 IPC 调用

阶段 5：可观测性 (增量)
  ├── 引入 OpenTelemetry
  ├── 关键链路加 span
  └── 统一结构化日志格式
```

---

## 六、阶段 0：性能基准测试（前置必做）

> **原则：先量化，再重构。** 任何性能重构必须以 pprof 火焰图的量化解剖为前提，拒绝盲改。

### 为什么先做这步

当前对性能瓶颈的判断（CGo 开销、Channel 切换、Opus 解码）都是**静态推断**，缺乏运行时数据支撑。贸然重构可能导致：

- 优化了 CGo 调用，但实际瓶颈在 Channel 阻塞
- 优化了 Opus 解码，但实际瓶颈在 ASR WebSocket 等待
- 改造了 ASR 并发池，但实际瓶颈在 LLM 首 token 延迟

pprof 能给出**精准的 CPU/内存分配占比**，让重构有的放矢。

### 执行步骤

#### 1. 启用 pprof

项目已内置 pprof 配置，只需开启：

```yaml
# config.local.yaml
server:
  pprof:
    enable: true
    port: 6060
```

启动服务后访问 `http://localhost:6060/debug/pprof/`。

#### 2. CPU Profile + 火焰图

**采集条件**：设备正常对话 5 分钟（含多次 listen start -> ASR -> LLM -> TTS -> listen stop 完整周期）

```sh
# 采集 5 分钟 CPU profile
go tool pprof -seconds=300 http://localhost:6060/debug/pprof/profile

# 在 pprof 交互中生成火焰图 SVG
(pprof) web

# 或命令行直接生成 SVG
go tool pprof -svg -seconds=300 http://localhost:6060/debug/pprof/profile > cpu_profile.svg
```

**重点观察**：
- `runtime.cgo` 系列函数占比 -- CGo 跨语言调用开销
- `internal/util.(*OpusRepacketizer).*` -- Opus 解码耗时
- `runtime.chanrecv` / `runtime.chansend` -- Channel 上下文切换开销
- `net/http.(*conn).serve` / `gorilla/websocket.*` -- 网络 I/O 等待
- `internal/pool.(*Manager).Acquire` -- 资源池获取耗时

#### 3. 内存 Profile

```sh
# 堆内存分配
go tool pprof http://localhost:6060/debug/pprof/heap

# 生成 SVG
go tool pprof -svg http://localhost:6060/debug/pprof/heap > heap_profile.svg

# 查看 alloc_space (累计分配)
go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap
```

**重点观察**：
- `internal/util` 音频帧分配频率 -- 是否有 GC 压力
- `internal/app/server/chat.(*ChatSession)` 会话对象内存占用
- `bytes.MakeCopy` / `runtime.makeslice` -- 音频帧拷贝开销

#### 4. Goroutine Profile

```sh
# 查看 goroutine 堆栈
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 生成 SVG
go tool pprof -svg http://localhost:6060/debug/pprof/goroutine > goroutine_profile.svg
```

**重点观察**：
- 每个 ASR 连接是否独立 goroutine，有无 goroutine 泄漏
- Channel 收发阻塞的 goroutine 数量
- `runtime.selectgo` 占比 -- 多路 select 的调度开销

#### 5. Block Profile（互斥/阻塞）

```sh
# 需先在代码中开启 block profiling
# runtime.SetBlockProfileRate(1)
go tool pprof http://localhost:6060/debug/pprof/block
```

**重点观察**：
- `internal/pool.(*Manager).Acquire` -- 资源池等待时间
- Channel 收发阻塞时间 -- TTS 队列、ASR 队列
- `sync.Mutex.Lock` -- 锁竞争热点

#### 6. Benchmark 测试

对关键路径编写 Benchmark，量化单次操作耗时：

```go
// internal/util/opus_repacketizer_test.go
func BenchmarkOpusRepacketize(b *testing.B) {
    repacketizer := NewOpusRepacketizer(...)
    frame := make([]byte, 960) // 模拟 opus 帧
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repacketizer.Repacketize(frame)
    }
}

// internal/pool/manager_test.go
func BenchmarkPoolAcquireRelease(b *testing.B) {
    manager := NewManager(...)
    manager.RegisterResourceType("asr", ...)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        wrapper, _ := manager.Acquire("asr", "test")
        manager.Release(wrapper)
    }
}

// internal/util/sentence_test.go
func BenchmarkSentenceSplit(b *testing.B) {
    text := "你好这是一个用于测试句子切分性能的文本..."
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SplitSentence(text)
    }
}
```

```sh
go test -bench=. -benchmem -count=5 ./internal/util/... ./internal/pool/...
```

### 产出物

完成阶段 0 后，应产出以下量化解剖报告：

| 产出物 | 内容 | 用途 |
|--------|------|------|
| `cpu_profile.svg` | CPU 火焰图 | 定位 CPU 热点函数 |
| `heap_profile.svg` | 堆内存火焰图 | 定位内存分配热点 |
| `goroutine_profile.svg` | Goroutine 堆栈 | 排查 goroutine 泄漏/阻塞 |
| `block_profile.svg` | 阻塞分析 | 定位锁/Channel 等待 |
| `benchmark_results.txt` | Benchmark 数据 | 量化单次操作耗时 |
| **瓶颈分析报告** | 汇总以上数据，给出优先级排序 | 指导阶段 3-4 重构 |

### 瓶颈判定标准

根据 pprof 数据决定后续重构优先级：

| 瓶颈位置 | CPU 占比 | 优先级 | 对应重构阶段 |
|---------|---------|--------|-------------|
| CGo 调用（Opus/TEN-VAD） | >15% | P0 | 阶段 4：CGo 解耦 |
| Channel 阻塞 | >10% | P1 | ASR 并发池重构 |
| 资源池 Acquire 等待 | >5% | P2 | pool/manager 优化 |
| GC 压力（音频帧分配） | >10% | P3 | 音频帧内存池 |
| 网络 I/O 等待 | 占比高但正常 | -- | 非重构项，改用非阻塞 I/O |

**只有当 pprof 数据确认瓶颈后，才启动对应阶段的重构。** 数据不足或瓶颈不明显时，优先做低风险的布局和配置重构（阶段 1-2）。

---

## 参考标准

| 标准 | 来源 |
|------|------|
| Go 项目布局 | [golang-standards/project-layout](https://github.com/golang-standards/project-layout) |
| 12-Factor App | [12factor.net](https://12factor.net/) |
| 六边形架构 | [Alistair Cockburn - Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) |
| Clean Architecture | [Robert C. Martin - Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) |
| Go Workspaces | [Go 1.20 Release Notes](https://go.dev/doc/go-workspace) |
| Effective Go | [go.dev/doc/effective_go](https://go.dev/doc/effective_go) |
