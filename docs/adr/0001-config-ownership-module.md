# 配置所有权：类型化 Settings + 窄 Reader 深模块

配置取值从散落在 19 个文件的 viper 全局字符串键，收进新建的 `internal/pkg/config` 深模块：一次解析得到类型化、不可变的 `Settings`，消费方只依赖按角色拆分的窄 `Config Reader`（DeviceConfigReader / SystemConfigReader / ActivationReader），`Config Source` 由 Init 构造一次并持有（连接复用，设备上下线不再每帧重建 manager HTTP client），12-Factor 环境变量作为解析链中的一个 adapter，来源优先级固定为 defaults → base 文件 → local overlay → env → remote merge，由单一 `Resolve()` 返回最终 Settings。

决策理由：现状 viper 全局导致键名与默认值散落、同一概念默认值不一致（config_provider.type 在 config.yaml 为 memory、config_init.go 为 redis）；宽 UserConfigProvider 接口 9 个方法让消费者被迫依赖全部；manager/redis 双生命周期并存。`internal/config` 孤儿模块删除，`internal/domain/config` 收窄为 Config Source 适配器。

- **Status**: accepted
- **Considered Options**:
  - 原地深化 `internal/domain/config` —— 否决：配置被 domain/app/cmd/pool/rag/memory 等横切消费，放 `pkg` 才能让依赖方向正确（domain 依赖 pkg 而非反向）
  - 保留单个宽 `ConfigReader` —— 否决：浅接缝，加方法即波及全部消费者
- **Consequences**: 热更 System Config 编排留在 `cmd/server`，作为独立深化项后续处理；词汇表已按 `CONTEXT.md` 固化 Config Source / Config Reader / Settings / System Config / Device Config
