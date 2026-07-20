# 🛠️ 项目工程化标准与重构方案

本指南立足于**现代 Go 语言企业级工程化标准（Production-Grade Standards）**和**清洁/六边形架构（Clean / Hexagonal Architecture）**，对 `xiaozhi-esp32-server-golang` 项目的代码布局、配置规划、功能划分进行深度解剖，评估其合理性、标准性，并给出具体的重构设计蓝图。

---

## 一、 📊 现状评估与痛点分析

### 1. 📂 项目布局（Layout）
*   **合理性：7.5 / 10**（功能模块化清晰，核心领域划分明确，具备了企业级项目的基本骨架）。
*   **标准性：5.5 / 10**（具有明显的“野蛮生长”痕迹，存在顶级目录污染、混杂的 Monorepo 管理等问题）。
*   **核心痛点**：
    1.  **顶级目录污染**：`logger/`、`lib/`、`storage/` 被直接置于项目根目录。在标准的 Go 工程中，除了构建、配置、文档和核心代码入口，根目录应当保持绝对干净。
    2.  **Monorepo 缺乏协同**：项目包含三个独立的 Go Module（根目录主服务、`manager/backend/`、Git 子模块 `asr_server/`），但未配置 **Go Workspaces (`go.work`)**，导致本地联合调试、代码跳转及依赖管理极为繁琐。
    3.  **内部目录语义混杂**：`internal/components/http/`、`internal/pkg/`、`internal/util/` 职责重叠。

### 2. ⚙️ 配置规划（Configuration）
*   **合理性：6.5 / 10**（支持多配置 Provider 模式：memory、manager、redis）。
*   **标准性：5.0 / 10**（覆盖逻辑、格式及敏感信息管理不标准）。
*   **核心痛点**：
    1.  **覆盖逻辑反直觉**：`config.local.yaml` 是对 `config.yaml` 的**完整替代**，而非**合并覆盖（Merge/Override）**。当默认配置新增项时，本地配置因缺项会报运行错误。
    2.  **配置格式与源碎片化**：项目并存 `config.yaml`、`mqtt_config.json` 以及 `asr_server/config.json`。
    3.  **违背 12-Factor 环境变量原则**：作为高并发后端，项目不支持直接通过 OS 环境变量注入敏感信息（如 `DEEPSEEK_API_KEY`、`MYSQL_PASSWORD`），这对容器化（Docker/K8s）部署极不友好。

### 3. 🧩 功能划分与分层（Functional Partitioning）
*   **合理性：7.0 / 10**（Provider 抽象层设计良好，方便对接多家 ASR/TTS 服务）。
*   **标准性：5.5 / 10**（框架实现泄露到领域层，CGo 强耦合导致编译笨重）。
*   **核心痛点**：
    1.  **具体框架向 Domain 泄露**：`internal/domain/llm/eino_llm/` 位于领域层（Domain）。Eino 是具体的 LLM 编排框架，属于**外层基础设施/适配器**（Infrastructure/Adapter）。领域层应当保持 100% 纯粹，不依赖任何第三方业务框架。
    2.  **CGo 编译屏障严重**：主程序通过 `asr_enabled` 构建标签与 CGo 代码耦合，导致整个主服务的交叉编译非常笨重，丧失了 Go 跨平台快速分发的优势。

---

## 二、 🛠️ 怎么改？（Reconstruction Plan）

### 1. 📁 布局重构：遵循 golang-standards 规范
*   **参照标准**：[golang-standards/project-layout](https://github.com/golang-standards/project-layout)
*   **重构方案**：
    *   **收拢顶级目录**：将根目录的 `logger/` 移动至 `internal/pkg/logger/`；将 `lib/` 移动至 `internal/pkg/lib/` 或 `build/lib/`；将 `storage/` 移动至 `data/storage/`。
    *   **启用 Go 1.20+ 工作区（Workspaces）**：在根目录下创建 `go.work`，统一管理三个独立 Go 模块。
        ```go
        go 1.20
        use (
            .
            ./manager/backend
            ./asr_server
        )
        ```
    *   **规范 `cmd/` 的职责**：`cmd/server/main.go` 只做：命令行参数解析、加载配置、优雅退出信号监听。其余复杂的初始化逻辑全部内聚到 `internal/app/`。

### 2. ⚙️ 配置重构：引入层次化合并与 12-Factor 标准
*   **参照标准**：[The Twelve-Factor App - III. Config](https://12factor.net/config)
*   **重构方案**：
    *   **多层级叠加覆盖（Merge）**：引入 Viper 或自定义配置 Loader，加载顺序为：
        $$\text{Default YAML} \longleftarrow \text{Local YAML (GitIgnored)} \longleftarrow \text{Environment Variables (OS)} \longleftarrow \text{Database Config}$$
        本地只需在 `config.local.yaml` 写修改过的 API Key 即可，无需完整复制。
    *   **绑定环境变量**：通过 `Viper.AutomaticEnv()` 或手写映射，自动将 OS 环境变量绑定到配置项，如：
        ```bash
        export XIAOZHI_LLM_DEEPSEEK_API_KEY="sk-xxxx"
        ```
        无缝支持容器化部署。

### 3. 🧩 架构重构：六边形架构（Ports and Adapters）
*   **参照标准**：**Hexagonal Architecture（端口与适配器模式）**
*   **重构方案**：
    *   **净化 Domain（领域层）**：`internal/domain/` 下只定义 **纯粹的 Interface（Ports/端口）** 和 **领域核心模型**。例如：
        *   `internal/domain/llm/provider.go` 定义核心流式接口 `LLMProvider`，不得 import 任何 Eino 包。
    *   **下放实现到 Infrastructure（适配器层）**：
        *   新建 `internal/infrastructure/llm/eino/`，实现上述 `LLMProvider` 接口，在此处 import Eino 框架及编排。
        *   新建 `internal/infrastructure/vad/webrtc/`，存放 WebRTC VAD 的实现。
    *   **彻底斩断 CGo 物理绑定**：
        *   主服务（纯 Go）取消 Tag 编译 CGo 依赖的代码。
        *   强制将 ASR/声纹服务（`asr_server`）作为一个**独立进程**部署。
        *   主服务通过 **gRPC / Local IPC Unix Socket / WebSocket** 与其交互。主服务获得 100% 纯 Go 极速编译分发体验，`asr_server` 独立在目标机器编译，完美解耦。

---

## 🗺️ 三、 重构后项目布局蓝图（Target Blueprint）

重构后的系统文件布局应当调整为如下结构：

```
xiaozhi-esp32-server-golang/
├── go.work                         # Go Workspaces（多 Module 联合调试）
├── go.mod                          # 主服务 Module
├── Makefile
├── README.md
│
├── cmd/                            # 唯一的执行入口
│   └── xiaozhi-server/
│       └── main.go                 # 命令行解析、配置加载、实例化并启动 App
│
├── config/                         # 纯静态、默认公用配置文件
│   └── default.yaml
│
├── data/                           # 本地持久化与运行时临时数据
│   ├── storage/                    # 存储声纹向量、SQLite 数据库等
│   └── uploads/                    # 存储声音复刻音频
│
├── internal/                       # 主服务核心（受 internal 保护，防止外部非法引用）
│   ├── app/
│   │   └── server/                 # 核心服务器初始化逻辑（服务、端口启动等）
│   │
│   ├── domain/                     # 【领域层】无外部框架依赖，定义协议、实体与端口
│   │   ├── session/                # 会话流转状态机、信令
│   │   ├── asr/                    # type ASR interface
│   │   ├── llm/                    # type LLM interface
│   │   ├── tts/                    # type TTS interface
│   │   └── vad/                    # type VAD interface
│   │
│   ├── infrastructure/             # 【基础设施/适配器层】具体框架与第三方服务的技术实现
│   │   ├── llm/
│   │   │   ├── eino/               # 封装 Eino 框架的 LLM 具体实现
│   │   │   └── mock/               # Mock 本地测试 LLM 实现
│   │   ├── vad/
│   │   │   ├── silero/             # Silero VAD 模型加载实现
│   │   │   └── webrtc/             # 纯 Go WebRTC VAD 实现
│   │   └── db/                     # GORM / SQLite 具体存储适配
│   │
│   └── pkg/                        # 通用、非业务公共库（全项目共享）
│       ├── logger/                 # 结构化日志组件
│       ├── pool/                   # 并发连接池/资源池
│       └── config_loader/          # 层次化配置加载器（支持 OS 环境变量合并）
│
├── manager/                        # 管理后台服务（子 Module）
│   ├── backend/                    # Go API 核心服务
│   └── frontend/                   # Vue 3 静态前端
│
├── asr_server/                     # CGo 独立 ASR/声纹微服务（子 Module，进程间 RPC 解耦）
│   └── main.go
│
└── test/                           # 自动化集成测试与高吞吐压力测试工具
```

---

## 四、 🌟 改造收益

1.  **闪电般的交叉编译**：主服务解开 CGo 绑定，成为 100% 纯 Go 代码。编译、多平台分发、CI/CD 构建耗时将由分钟级下降至秒级。
2.  **优雅、流畅的开发体验**：启用 `go.work` 后，IDE 能够完美感知跨 Module 跳转，不需要再通过手动替换 `replace` 依赖路径来联调。
3.  **零安全配置隐患**：彻底隔离敏感 Token。在生产环境部署时直接以 Docker 环境变量的形式（12-Factor）安全注入，规避本地 `.yaml` 提交代码库泄露敏感信息的风险。
4.  **清晰可测的代码逻辑**：领域层只关注纯业务逻辑，不关注用的是 Eino 还是原生 SDK，单元测试（Unit Testing）编写耗时缩短 50% 以上。
