# Xiaozhi ESP32 Server

AI 语音助手的后端服务上下文，覆盖设备接入、对话链路、配置分发与部署形态。

## Language

**Device Config**:
单台设备生效的 AI 能力配置（ASR/LLM/TTS/VAD 等），由 Config Source 按设备维度拉取与隔离。
_Avoid_: 用户配置, user config

**System Config**:
服务级运行配置，可由管理后台经 API/WebSocket 热更新，影响全部连接而非单台设备。
_Avoid_: server config

**Config Source**:
Device Config 的取用来源抽象，现有三种实现：memory / manager / redis。
_Avoid_: provider（已被 AI Provider 占用）, 配置提供方

**Config Reader**:
消费方依赖的窄只读接口，从 Config Source 之上取配置，模块不直接触碰 viper 全局。
_Avoid_: 全局 viper 读取, GetConfig

**Settings**:
配置经一次解析后的类型化、不可变视图，供各模块只读消费。
_Avoid_: config（三义歧义）

**AI Provider**:
具体的外部 AI 能力实现（LLM/TTS/ASR 供应商），与配置取用概念区分。
_Avoid_: model（与模型文件混淆）
