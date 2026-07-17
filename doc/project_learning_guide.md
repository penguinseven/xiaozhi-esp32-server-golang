# 项目学习指南

> 面向已跑通基础对话的开发者，结合 `doc/` 和 `ai_doc/` 文档，系统化剖析项目架构与代码。

## 一、整体架构

```
┌─────────────────────────────────────────────────┐
│                   ESP32 设备                      │
│         (xiaozhi-esp32 固件)                      │
└──────┬──────────────────────────────┬───────────┘
       │ WebSocket (8989)              │ MQTT+UDP (2883/8990)
       │                              │
┌──────▼──────────────────────────────▼───────────┐
│              主服务 (cmd/server)                   │
│  ┌─────────────────────────────────────────┐     │
│  │  chat.go     消息路由 / 连接管理          │     │
│  │  session.go  会话主循环 (VAD->ASR->LLM->TTS) │  │
│  │  pool/       资源池管理                   │     │
│  │  config/     配置 provider (memory/manager) │     │
│  └─────────────────────────────────────────┘     │
│  ┌─────────────────────────────────────────┐     │
│  │  ASR  (doubao/funasr/aliyun/xunfei)     │     │
│  │  LLM  (deepseek/qwen/doubao/chatglm)    │     │
│  │  TTS  (edge/doubao_ws/cosyvoice/xunfei) │     │
│  │  VAD  (ten_vad/silero/webrtc)           │     │
│  └─────────────────────────────────────────┘     │
└──────────────────┬───────────────────────────────┘
                   │ HTTP API
┌──────────────────▼───────────────────────────────┐
│          管理后台 (manager/backend)               │
│  设备管理 / 配置管理 / 知识库 / 声纹 / 声音复刻    │
└──────────────────┬───────────────────────────────┘
                   │ Vue3 + Element Plus
┌──────────────────▼───────────────────────────────┐
│          管理前端 (manager/frontend)              │
└──────────────────────────────────────────────────┘
```

### 四个子系统

| 子系统 | 路径 | 职责 |
|--------|------|------|
| 主服务 | `cmd/server/` | WebSocket/MQTT/UDP 接入，ASR->LLM->TTS 全链路 |
| 管理后台后端 | `manager/backend/` | 设备/配置/知识库/声纹 REST API |
| 管理前端 | `manager/frontend/` | Vue3 + Element Plus 管理界面 |
| ASR 子模块 | `asr_server/` | sherpa-onnx + onnxruntime CGo 绑定 |

### 三种部署方式

| 方式 | 命令 | 适用场景 |
|------|------|----------|
| AIO 一体化 | `go build -tags "manager asr_server embed_ui"` | 单机快速部署 |
| 分离部署 | 各子系统独立编译运行 | 生产环境 |
| 单二进制 | 仅主服务 `go build` | 开发调试 |

参考：[architecture_overview.md](architecture_overview.md)、[compile_deploy.md](compile_deploy.md)

## 二、两条传输链路

设备到服务端有两条路径：

### 路径 1：WebSocket

- 设备直连 `ws://host:8989/xiaozhi/v1/`
- 单一 TCP 连接，信令和音频都走 WebSocket
- 参考：[websocket_connection_flow.md](websocket_connection_flow.md)

### 路径 2：MQTT + UDP

- **MQTT 传信令**：hello / listen start / listen stop 等
- **UDP 传音频**：opus 编码的音频帧
- 设备先 MQTT 连接 -> 发 hello 协商参数 -> 切到 UDP 收发音频
- 参考：[mqtt_udp.md](mqtt_udp.md)、[mqtt_udp_protocol.md](mqtt_udp_protocol.md)

### MQTT Topic 映射规则

```
/p2p/device_public/{mac}    设备 -> 服务端 (信令)
/p2p/device_sub/{mac}       服务端 -> 设备 (信令)
device-server               设备 -> 服务端 (hello)
/p2p/device_public/_server/lifecycle  生命周期事件
```

### MQTT 认证

- Client ID 格式：`GID_test@@@{deviceId}@@@{uuid}`
- Password 使用 HMAC-SHA256 签名
- 参考：[ota_mqtt_auth.md](ota_mqtt_auth.md)

## 三、会话主循环（核心）

整个项目最重要的链路，设备每次说话都走这条路径：

```
设备发送 listen start
  │
  ▼
session.go: OnListenStart()
  ├── 创建 VAD 资源 (pool 获取)
  ├── 创建 ASR 资源 (pool 获取)
  └── 启动 ASR 流式识别
  │
  ▼ (设备持续发送音频帧)
  │
  ├── VAD 检测到语音段 -> 送入 ASR
  ├── ASR 中间结果 (部分文字) -> 可用于打断判断
  └── ASR 最终结果 (完整文字) -> 送入 LLM
  │
  ▼
llm/eino_llm.go: Chat()
  ├── 组装 system_prompt + 历史 + 当前输入
  ├── 流式调用 LLM API
  └── 每个 chunk -> 分句 -> 送入 TTS 队列
  │
  ▼
tts/edge/edge.go: TextToSpeechStream()
  ├── 文本 -> 音频帧 (流式)
  └── 音频帧 -> 发回设备播放
  │
  ▼
listen stop -> 会话结束
```

### 关键代码入口

| 步骤 | 文件位置 | 函数 |
|------|----------|------|
| 消息路由 | `internal/app/server/chat/chat.go` | `收到文本消息` (L370) |
| 会话启动 | `internal/app/server/chat/session.go` | `OnListenStart` (L1235) |
| ASR 重启 | `internal/app/server/chat/session.go` | `OnListenStart start` (L1265) |
| LLM 调用 | `internal/domain/llm/eino_llm/eino_llm.go` | `Chat()` |
| TTS 合成 | `internal/domain/tts/edge/edge.go` | `TextToSpeechStream()` |

### 打断模式 (realtime_mode)

| 值 | 模式 | 说明 |
|----|------|------|
| 1 | VAD 打断 | VAD 检测到新语音段时打断当前 TTS |
| 2 | ASR 打断 | ASR 识别到文字时打断 |
| 3 | 声纹打断 | ASR 识别到特定声纹时打断 |
| 4 | ASR 结果打断 | ASR 出结果即打断（兼容流式或离线） |

### 延迟参考

funasr + qwen2.5-72b + cosyvoice 组合可实现 **1-1.3 秒**内回复（ASR->LLM->TTS 首帧约 975-1394ms）。参考：[delay_test.md](delay_test.md)

## 四、配置体系

### 三种 Config Provider 模式

| 模式 | 配置来源 | 适用场景 |
|------|----------|----------|
| `memory` | 直接读 `config.yaml` / `config.local.yaml` | 开发调试，单机 |
| `manager` | 从管理后台数据库读 | 多设备管理，生产环境 |
| `redis` | 从 Redis 读 | 分布式部署 |

### 配置优先级

```
config.local.yaml (完整替代，非合并)
  > config.yaml
```

### 配置覆盖机制

`config.local.yaml` 是 `config.yaml` 的完整替代（不是合并），用于本地开发覆盖敏感配置（API Key 等）。参考：[config.md](config.md)

### Manager 模式配置

使用 `manager` 模式时，需要在 Web 后台配置默认的 LLM/TTS/ASR/VAD 配置：

1. 登录管理后台 `http://127.0.0.1:8080`
2. 完成 5 步配置向导：OTA -> VAD -> ASR -> LLM -> TTS
3. 每类配置需设置 `is_default = true` 才会被默认使用

参考：[manager_console_guide.md](manager_console_guide.md)

## 五、代码阅读路径

### 推荐阅读顺序

```
入口层
  cmd/server/main.go              程序入口

核心层 (internal/domain/)
  asr/base.go                    ASR Provider 模式 (70行)
  tts/edge/edge.go               最简单的 TTS (200行)
  llm/eino_llm/eino_llm.go       LLM 流式对话 + Tool Calling
  chat/session.go                会话主循环 (核心，最复杂)

支撑层 (internal/)
  pool/manager.go                 资源池管理
  config/memory/memory.go        Memory 配置加载
  config/manager/                 Manager 配置加载

管理层 (manager/)
  backend/controllers/            REST API
  frontend/src/                   Vue3 界面
```

### 源码搭建

参考：[source_code_setup_tutorial.md](source_code_setup_tutorial.md)

macOS 关键依赖：
- Opus：`brew install opus libopusfile pkg-config`
- OnnxRuntime：下载 `libonnxruntime.1.21.0.dylib` 到 `lib/ten-vad/lib/macOS/`
- Build Tags：`asr_enabled` / `!asr_enabled` 控制 ASR 编译

## 六、Provider 体系

### ASR 引擎

| Provider | 说明 | 需要 |
|----------|------|------|
| `funasr` | FunASR WebSocket | 本地 FunASR 服务 |
| `aliyun_funasr` | 阿里云 DashScope | API Key |
| `doubao` | 豆包 SAUC | AppID + Access Token |
| `aliyun_qwen3` | 阿里云 Qwen3 | API Key |
| `xunfei` | 讯飞 IAT | AppID + API Key + Secret |

### LLM 引擎

| Provider | 说明 | 需要 |
|----------|------|------|
| `deepseek` | DeepSeek V3 (硅基流动) | API Key |
| `deepseek_v4_flash` | DeepSeek V4 Flash (官方) | API Key |
| `doubao_deepseek` | 豆包 DeepSeek (火山引擎) | Ark API Key |
| `qwen_72b` | 通义千问 (硅基流动) | API Key |
| `chatglmllm` | 智谱 GLM-4-Flash | API Key |

### TTS 引擎

| Provider | 说明 | 需要 |
|----------|------|------|
| `edge` | 微软 Edge TTS (免费) | 无 |
| `doubao_ws` | 豆包 WebSocket TTS | AppID + Token + Voice |
| `cosyvoice` | CosyVoice | API URL |
| `xunfei` | 讯飞在线 TTS | AppID + Key + Secret |
| `xiaozhi` | 小智官方 TTS | 设备 ID + Token |

### VAD 引擎

| Provider | 说明 | 需要 |
|----------|------|------|
| `ten_vad` | TEN-VAD (onnxruntime) | 模型文件 |
| `silero_vad` | Silero VAD | onnx 模型 |
| `webrtc_vad` | WebRTC VAD | 无 |

## 七、进阶功能

### 1. MCP 工具调用

让 AI 能调用外部工具（查天气、控制设备、查知识库）。

- 基于 Eino 框架实现通用工具管理
- 支持全局 SSE 连接多 MCP 服务器
- 支持设备维度 WebSocket 连接的独立工具管理
- 参考：[mcp.md](mcp.md)
- 设计演进：[ai_doc/chat_hook_plugin_v2_design.md](../ai_doc/chat_hook_plugin_v2_design.md)

### 2. 知识库 (RAG)

让 AI 基于私有文档回答问题。

- 三层架构：管理员配置检索 provider -> 用户创建知识库 -> 智能体关联知识库
- 支持 Dify / RAGFlow / WeKnora
- 对话时通过 `search_knowledge` 工具触发检索
- 参考：[knowledge_base.md](knowledge_base.md)
- 增强方案：[ai_doc/knowledge_mcp_auto_trigger_plan.md](../ai_doc/knowledge_mcp_auto_trigger_plan.md)

### 3. 声纹识别

识别说话人，自动切换 TTS 音色。

- sherpa-onnx 提取 192 维声纹 embedding
- Qdrant 向量库存储与检索
- 支持 注册/识别/验证/流式 WebSocket 识别
- 参考：[speaker_identification.md](speaker_identification.md)

### 4. 声音复刻

用用户自己的声音做 TTS。

- 支持 minimax / cosyvoice / aliyun_qwen
- 用户上传音频或浏览器录音
- 管理员可控制复刻额度
- 参考：[voice_clone.md](voice_clone.md)

### 5. 视觉识别

让设备能拍照识别。

- 调用外部视觉服务（阿里云 Qwen-VL、火山豆包 Vision）
- LLM 识别拍照意图 -> MCP Tool 下发拍照指令 -> 终端识别返回
- 参考：[vision.md](vision.md)

### 6. OpenClaw Agent

可编程的对话代理。

- 按 agent_id 管理 WebSocket 连接池
- 关键词触发进入/退出 OpenClaw 模式
- 模式内消息绕过 LLM 直接走 OpenClaw 再走 TTS
- 设备离线时消息入内存队列，上线补发
- 参考：[ai_doc/openclaw_agent_integration_plan.md](../ai_doc/openclaw_agent_integration_plan.md)

## 八、设计文档 (ai_doc/)

`ai_doc/` 是架构演进方案，不是已实现的功能文档，适合理解项目设计思路：

| 文档 | 核心内容 |
|------|----------|
| [mqtt_lifecycle_transport_plan.md](../ai_doc/mqtt_lifecycle_transport_plan.md) | MQTT 生命周期驱动 Transport 预创建，设备上线即预热 |
| [chat_session_lazy_reuse_refactor_plan.md](../ai_doc/chat_session_lazy_reuse_refactor_plan.md) | ChatSession 懒创建+复用，hello 拆为 transport 级和 chat 级 |
| [llm_interrupt_extra_plan.md](../ai_doc/llm_interrupt_extra_plan.md) | 打断时在历史消息标记 `[用户打断]`，让模型理解被打断的上下文 |
| [chat_hook_plugin_v2_design.md](../ai_doc/chat_hook_plugin_v2_design.md) | Hook 体系演进为 Interceptor + Observer 框架 |
| [knowledge_mcp_auto_trigger_plan.md](../ai_doc/knowledge_mcp_auto_trigger_plan.md) | 知识库无感触发检索，支持定向 knowledge_base_ids |
| [openclaw_agent_integration_plan.md](../ai_doc/openclaw_agent_integration_plan.md) | OpenClaw Agent 维度集成，关键词触发+离线补发 |
| [weknora_integration_plan.md](../ai_doc/weknora_integration_plan.md) | WeKnora 知识库检索 provider 集成 |
| [branch_review_findings_20260412.md](../ai_doc/branch_review_findings_20260412.md) | 代码审查记录，P1/P2 问题 |

## 九、压测与调试

### 压测工具

```sh
# WebSocket 并发压测
docker run ws_multi -count 10 -server ws://host:8989 -text "你好"
```

参考：[websocket_meter.md](websocket_meter.md)、[fullchain_mock_pressure_test_plan.md](fullchain_mock_pressure_test_plan.md)

### 外部 AI Mock

开发时可用 Mock 服务替代真实 ASR/LLM/TTS：参考 [mock_external_ai_server.md](mock_external_ai_server.md)

### 管理后台测试

内置 VAD/ASR/LLM/TTS/OTA 五类可视化测试，在后台 `配置管理` 页操作。

## 十、学习路径建议

```
第一步：验证完整链路
  ├── 设备连接 -> ASR -> LLM -> TTS -> 设备播放
  └── 确认日志中每一步都有输出

第二步：理解数据流
  ├── 读 chat.go 消息路由
  ├── 读 session.go 会话主循环
  └── 跟着 OnListenStart -> handleLLMResponse -> handleTts 走一遍

第三步：动手改
  ├── 换 LLM 模型，对比回复风格
  ├── 换 TTS 音色
  ├── 改 system_prompt 角色人设
  ├── 调 realtime_mode 打断策略
  └── 加自定义 MCP 工具

第四步：深入核心模块
  ├── session.go (最核心，2000+ 行)
  ├── asr/base.go (Provider 模式，70 行)
  ├── tts/edge/edge.go (最简单 TTS，200 行)
  ├── llm/eino_llm.go (LLM 流式 + Tool Calling)
  └── pool/manager.go (资源池设计)

第五步：进阶功能
  ├── MCP 工具调用
  ├── 知识库 RAG
  ├── 声纹识别
  ├── 声音复刻
  ├── 视觉识别
  └── OpenClaw Agent
```

## 十一、传输层深入理解

### 三个层次，互不混淆

```
层次1: 传输层 (设备怎么连服务端)     ← MQTT / WebSocket / UDP
层次2: 业务层 (设备连上后做什么)     ← 听音 -> 识别 -> 对话 -> 播报
层次3: 能力层 (具体用哪家AI服务)     ← ASR / LLM / TTS / VAD 配置
```

**传输层和能力层完全独立**。换 ASR provider 不需要改 MQTT，换传输方式也不需要改 LLM。

### 两条传输链路

设备到服务端有两条路径，**二选一，不是同时用**：

#### 路径 1：WebSocket

- 设备直连 `ws://host:8989/xiaozhi/v1/`
- 单一 TCP 连接，信令和音频都走 WebSocket
- 简单，但音频走 TCP 会因丢帧重传导致延迟累积
- 适合简单部署

#### 路径 2：MQTT + UDP

- **MQTT 传信令**：hello / listen start / listen stop 等，走 TCP 保证可靠
- **UDP 传音频**：opus 编码的音频帧，走 UDP 保证低延迟
- 复杂，但音频丢几帧不影响语音质量，延迟更低
- 适合追求低延迟的语音场景

| | WebSocket 路径 | MQTT + UDP 路径 |
|---|---|---|
| 连接数 | 1 条 | 2 条 (MQTT + UDP) |
| 信令传输 | WebSocket | MQTT |
| 音频传输 | WebSocket | UDP |
| 延迟 | 稍高 (TCP 可靠传输) | 更低 (UDP 不保证顺序) |
| 可靠性 | 高 | 信令可靠，音频可能丢帧 |
| 复杂度 | 简单 | 复杂 (要协调两条连接) |
| 适用场景 | 简单部署 | 追求低延迟的语音场景 |

### 为什么不能单用 UDP 或单用 MQTT

- **单用 UDP**：信令可能丢失/乱序，`listen stop` 比 `listen start` 先到会导致逻辑混乱
- **单用 MQTT**：音频走 TCP，一帧丢了后面全部排队等待重传，延迟累积越来越大

信令必须可靠（用 TCP/MQTT），音频宁可丢也不要等（用 UDP），各取所长。

### MQTT Server 是项目内置的

项目**内置了 MQTT 服务器**，不需要安装 mosquitto 或 EMQX：

```yaml
mqtt_server:
  enable: true           # 内置 MQTT 服务器
  listen_port: 2883      # 设备连这个端口
  enable_auth: false     # 默认不开认证
```

也支持连接外部 MQTT Broker（如 EMQX），只需把 `mqtt.broker` 改成外部地址。

### OTA 是"地址簿"

OTA = Over The Air，不是升级固件，而是**设备开机时问服务端要连接地址**：

```
设备开机
  │
  ▼  HTTP POST -> /xiaozhi/ota/
服务端返回:
  {
    "websocket": { "url": "ws://192.168.8.114:8989/xiaozhi/v1/" },
    "mqtt": { "enable": true, "endpoint": "192.168.8.114:2883" }
  }
  │
  ▼  设备根据 mqtt.enable 选择路径
  ├── mqtt.enable = true  -> 走 MQTT+UDP
  └── mqtt.enable = false -> 走 WebSocket
```

固件里只写死 OTA 地址，不写死通信地址，方便换服务器和切换传输方式。

## 十二、信令体系

### 什么是信令

信令是**控制命令**，告诉对方"我要干什么"。和音频（实际声音数据）分开传输。

```
信令 = 指挥命令 (文本，几十个字节)
音频 = 实际内容 (声音，每秒几千字节)
```

### 信令一览表

| 信令 | 方向 | 含义 | 不发会怎样 |
|------|------|------|-----------|
| `hello` | 设备->服务端 | 握手，协商音频格式 | 服务端不知道用什么格式解码音频 |
| `listen start` | 设备->服务端 | 我要说话了，启动 ASR | 服务端不会识别你的话 |
| `listen stop` | 设备->服务端 | 我说完了，开始回复 | 服务端不知道何时开始 LLM 回复 |
| `tts start` | 服务端->设备 | 开始播放语音 | 设备不知道何时开始播放 |
| `tts stop` | 服务端->设备 | 停止播放 | 用户打断时声音不会停 |
| `speak_request` | 服务端->设备 | 我要主动跟你说话 | 服务端无法主动发起对话 |
| `speak_ready` | 设备->服务端 | 我准备好了，开始吧 | 服务端不知道设备是否就绪 |
| `goodbye` | 设备->服务端 | 断开连接 | |

### 正常对话信令流程

```
设备                                    服务端
  │
  ├── hello ────────────────────────->   协商音频参数
  │
  ├── listen start ─────────────────->   启动 VAD + ASR
  │
  ├── [音频帧][音频帧] ──────────────>   VAD -> ASR -> LLM -> TTS
  │
  ├── listen stop ───────────────────>   开始 LLM 回复
  │
  │ <── tts start + [音频帧] ────────    播放回复
  │
  │ <── tts stop ───────────────────     播放完毕
```

### speak_request：服务端主动发起

`speak_request` 是**服务端发给设备的信令**，用于服务端主动发起对话：

```
服务端                                    设备
  │
  ├── MQTT: speak_request ──────────->   我要主动跟你说话
  │   {session_id, text, auto_listen}     设备从空闲状态唤醒
  │
  │ <── MQTT: speak_ready ─────────────   我准备好了
  │
  ├── UDP: TTS 音频帧 ───────────────>   播放语音
  │
  │ <── MQTT: listen start ────────────   (如果 auto_listen=true)
  │                                       播完后自动进入聆听
```

**为什么 speak_request 主要用于 MQTT 协议？**

WebSocket 空闲时会断开连接，服务端找不到设备发不了消息。MQTT 保持长连接，设备空闲时连接还在，服务端随时能主动推送。

相关配置：

```yaml
chat:
  speak_request_reuse_window_ms: 60000  # UDP 链路复用窗口（毫秒）
```

设备空闲后 60 秒内 UDP 链路还在，服务端发 `speak_request` 可以直接复用；超过 60 秒要重新 hello 协商。
