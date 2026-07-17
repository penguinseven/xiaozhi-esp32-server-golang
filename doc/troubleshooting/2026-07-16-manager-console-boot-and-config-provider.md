# 2026-07-16 管理后台 config_provider=manager 模式下的一连串启动/配置疑难

> 场景：`config_provider.type=manager`，主服务 `cmd/server` 与管理后台 `manager/backend` 同工作树，SQLite (`data/xiaozhi.db`) 作为默认存储。设备通过 MQTT/UDP + WebSocket 接入，前端仪表板显示"在线设备"。
>
> 这一天陆续排掉了：配置向导脏数据、8989 端口占用、LLM 404、"设备已连接但不在线"、GORM record-not-found 噪音、OTA `www.tb263.cn` 占位符等。

---

## 1. 配置向导写入 broker/external_host 脏数据（根因链）

### 现象
- `mqtt_wizard_default.broker` = `192.168.8.114:8989`
- `udp_wizard_default.external_host` = `192.168.8.114:8989`
- `ota_ota_config.test.mqtt.endpoint` = `192.168.8.114:8989:1883`
- 主服务日志出现 `too many colons in address` 或 MQTT 客户端连的是 web 端口。

### 根因
`manager/frontend/src/views/admin/ConfigWizard.vue` 的 OTA 步骤"域名或 IP"输入框直接透传用户输入到 `broker / external_host / mqtt.endpoint`：

- `finalMqttEndpoint = otaForm.host.trim() + ':' + otaForm.mqttServerPort`
- `saveMqttConfig() / saveUdpConfig() / saveOta()` 中的 host 直接来自 `otaForm.host.trim()`。

如果 host 里填 `192.168.8.114:8989`（很自然的写法），web 端口 8989 就渗透到 MQTT/UDP 字段。

### 结论
- 本轮不改前端，只在 SQLite 里清脏数据（见下）；后续可考虑在向导里做 `hostOnly()` 净化。

---

## 2. `WebSocket 服务器启动失败: listen tcp 0.0.0.0:8989: bind: address already in use`

### 现象
- 主服务启动后 fatal 退出：`internal/app/server/websocket/websocket_server.go:125`。

### 根因（不是外部进程）
- `config_provider.type=manager` 时，`cmd/server/config.go:170 updateConfigFromAPI` 每 5 分钟拉后台 `system_configs` 并合并进 viper。
- SQLite 里 `mqtt_server_mqtt_server_config` 被向导写成 `listen_port=8989`。
- `internal/app/mqtt_server/mqtt_server.go:71-79` 读 `mqtt_server.listen_host/listen_port` 起 TCP listener，先占住 8989。
- 紧接着 `WebSocketServer.Start()` 再 bind 8989，失败退出。

### 修复
- SQLite 里 `DELETE FROM configs WHERE config_id='mqtt_server_mqtt_server_config';`；本地 `mqtt_server.listen_port: 2883` 生效。
- 顺带在 `manager/backend/controllers/admin.go:getSystemConfigsData` 的 `getSelectedConfig` 里加 `Enabled` 过滤（见 §5），防止后台"禁用"配置继续被 push。

---

## 3. `连接MQTT服务器失败: dial tcp 192.168.8.114:1883: connection refused`

### 现象
- 主服务反复重连远端 MQTT，本地 yaml 明明写的是 `broker: 127.0.0.1, port: 2883`。
- 出处：`internal/app/server/mqtt_udp/mqtt_udp_adapter.go:229`。

### 根因
- `internal/app/server/app.go:117 currentMqttConfig()` 从 viper 读 `mqtt.*`。
- `mqtt_wizard_default` 被合并进 viper 后覆盖 yaml：`{broker:192.168.8.114, port:1883}`。
- 该行 DB `enabled=0` 却仍被下发（见 §5 bug）。

### 修复
- SQLite 中删除 `mqtt_wizard_default` / `udp_wizard_default`；MQTT 客户端与 UDP 回落到 yaml。
- 修 `getSelectedConfig`（§5）杜绝根因复发。

---

## 4. LLM 404: `deepseek-v4-flash does not exist or you do not have access`

### 现象
- `internal/domain/llm/eino_llm/eino_llm.go:321/:325` 反复报 404，Request id: `02178412...`。
- `curl https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions` 直连正常。

### 根因
- SQLite `llm_doubao_deepseek_default.json_data` 只写了 `{model_name, api_key, max_tokens, temperature, top_p}`，没有 `base_url` / `type`。
- `internal/domain/llm/base.go:88 resolveDefaultBaseURL("doubao")` 兜底填 `https://ark.cn-beijing.volces.com/api/v3`，而 `deepseek-v4-flash` 只存在于 `/api/coding/v3`。普通端点必 404。

### 修复
- 更新该行 `json_data` 严格对齐 `config/config.local.yaml:331-336` 的 `doubao_deepseek`：`type=openai` + `base_url=https://ark.cn-beijing.volces.com/api/coding/v3` + `model_name=deepseek-v4-flash` + `api_key` + `max_tokens=500`。
- 之前已经把 `resolveDefaultBaseURL` 改成"仅在用户未配置 base_url 时兜底"，此次事件确认了该修复的必要性。

---

## 5. `getSelectedConfig` 忽略 `Enabled` 字段

### 现象
- 后台把 `enabled=0` 的记录也当默认配置 push 给主服务，导致 §2/§3 反复复发。

### 根因
- `manager/backend/controllers/admin.go:getSystemConfigsData` 的旧实现只按 `IsDefault` 选，不看 `Enabled`。

### 修复
- 改成"必须 Enabled，且默认优先，否则取第一条 Enabled；全禁用返回 nil"。
- 上层写 `response["mqtt"/"udp"/"ota"/"auth"/"chat"/"local_mcp"]` 前统一做 `nil` 判断，避免把 nil 塞进响应。
- MCP 分支的选择也换成同一套逻辑。
- `manager/backend` 侧 `go build ./...` 通过（GOCACHE 借用 `/tmp` 绕开权限）。

---

## 6. GORM 日志误报 `record not found`

### 现象
- 大量 `[error] .../chat_history.go:90 record not found`。
- 业务逻辑本就是"先查后创建"，是正常路径。

### 根因
- GORM 默认 logger 把 `ErrRecordNotFound` 当作 error 打印。

### 修复
- `manager/backend/database/database.go` 的 `Init()` 里覆盖 logger：`gormlogger.New(..., Config{IgnoreRecordNotFoundError: true, LogLevel: Warn, SlowThreshold: 200ms})`。
- 之前在 `database.go` 上做过类似改动、被手动还原过；这次是最小侵入版本，只影响日志输出，不改业务行为。

---

## 7. `recv cmd error / recv audio error: context canceled`

### 现象
- 管理后台点"OTA 测试"后主服务日志红字：`[error] recv cmd error: context canceled` / `[error] recv audio error: context canceled`。

### 根因
- 后台伪造设备 `ota-test-device`（`manager/backend/controllers/admin.go:1554`）dial WebSocket，成功后立刻 `conn.Close()`。上游 ctx 被 cancel，主服务里的 `RecvCmd/RecvAudio` 收到 `context.Canceled`，属于正常关闭。

### 修复
- `internal/app/server/chat/chat.go`：`cmdMessageLoop` / `audioMessageLoop` 收到 `context.Canceled / io.EOF / context.DeadlineExceeded` 时降到 `debug` 并直接 return；只有真 IO 错误才走 `Errorf` + 重试计数。

---

## 8. OTA 配置里 `www.tb263.cn` 占位域名

### 现象
- `configs` 表里 `ota_ota_config.external.websocket.url` = `wss://www.tb263.cn:55555/go_ws/xiaozhi/v1/`，`external.mqtt.endpoint` = `www.youdomain.cn`。
- 设备拿到该 external 会连到无关公网域名。

### 修复
- SQLite 直接 UPDATE：`test` / `external` 全部指向局域网 `192.168.8.114`（`ws://192.168.8.114:8989/xiaozhi/v1/` + `mqtt endpoint=192.168.8.114:2883, enable=true`）。
- 未来上公网时把 `external` 换成真实域名即可，`test` 保留局域网调试。

---

## 9. SQLite 重置与最小 seed

### 场景
- 反复被向导脏数据、`enabled=0 却下发` 等问题折腾之后，选择把 `data/xiaozhi.db` 重命名为 `xiaozhi.db.reset-<ts>`（不删除，方便回滚）。
- backend 首次运行会 `AutoMigrate` 建表。

### 推荐 seed（只种 vad / asr / llm / tts / memory / ota）
严格对齐 `config/config.local.yaml`：

- `vad_ten_default` (ten_vad)
- `asr_doubao_default` (doubao；`ws_url=.../bigmodel_nostream`)
- `llm_doubao_deepseek_default` (含 `type=openai` + `base_url=.../api/coding/v3`，避免 §4 复发)
- `tts_edge_default`
- `memory_nomemo_default`
- `ota_ota_config` (test/external 都指向本机 `192.168.8.114:8989 / 2883`)

**不**种 mqtt / mqtt_server / udp / mcp 等，让主服务直接吃本地 yaml，避免任何时候脏数据再次污染。

---

## 10. "设备已连接但在线设备 = 0"

### 现象
- 主服务日志显示设备 WebSocket 已连接、`OnListenStart` 等回调走过；后台 Dashboard / Agents 列表却显示在线设备 0。

### 判定链
- 上报：`internal/app/server/app.go:475 DeviceOnline` → `provider.NotifyDeviceEvent`；`config_provider=manager` 时走 `internal/domain/config/manager/manager.go` 里的 WebSocket 客户端发 `POST /api/device/active`。
- 后台处理：`manager/backend/controllers/websocket.go:415 handleDeviceActiveRequest`，`UPDATE devices SET last_active_at=NOW() WHERE device_name = ?`。**匹配字段是 device_name，不是 device_code**。
- 计数：`manager/backend/controllers/user.go:898/:906` `WHERE last_active_at > NOW()-5min`。
- 前端：`manager/frontend/src/views/user/Agents.vue:250` `Date.now()-last_active_at < 5min`。

### 常见断点
1. `handleDeviceActiveRequest` 有没有跑到（日志 `处理设备活跃时间更新请求, device_id: xxx`）。
2. `device_name` 是否命中（若没命中会打 `设备不存在: xxx`）。
3. 上报 `device_id` 与 DB 里 `device_name` 大小写/分隔符是否一致（MAC 大小写、`:` vs `_` 是常见差异）。

### 临时验证
- 手动把 `last_active_at` 拨到当前：`UPDATE devices SET last_active_at=datetime('now','localtime') WHERE id=1;`，若前端立刻显示在线，说明问题在"上报路径 vs device_name 匹配"。

### 可选加固（未实施）
- 后台 `handleDeviceActiveRequest` 增加 `device_name=? OR device_code=?` 兜底；命中失败降级到 warn 并输出实际 device_id 便于排查。

---

## 11. 历史设备重回新库

### 场景
- 重置后新库 `configs` 空、`devices` 空，之前的一台设备 `34:cd:b0:a6:fb:8c` (nick_name=xiaozhi) 需要保留。

### 步骤
- 备份当前 DB 为 `xiaozhi.db.bak-preinsert-<ts>`。
- 从 `data/xiaozhi.db.reset-<ts>` ATTACH 后 `INSERT INTO devices SELECT ... FROM src.devices;`，保留 id/user_id/agent_id/nick_name/device_name/challenge/pre_secret_key/activated/last_active_at/created_at/updated_at。
- 由于 `device_code` 为空，`idx_devices_device_code` 唯一约束不会冲突。

---

## 常用备份/回滚命令

```bash
# 备份
cp data/xiaozhi.db data/xiaozhi.db.bak-$(date +%Y%m%d-%H%M%S)

# 查 configs 摘要
sqlite3 -header -column data/xiaozhi.db "SELECT type, config_id, provider, enabled, is_default FROM configs ORDER BY type, id;"

# 查设备匹配情况
sqlite3 -header -column data/xiaozhi.db "SELECT id, user_id, device_name, device_code, last_active_at FROM devices;"

# 8989 端口占用排查
lsof -nP -iTCP:8989 -sTCP:LISTEN
kill $(lsof -tnP -iTCP:8989 -sTCP:LISTEN)
```

---

## 教训 / 后续 TODO

- 配置向导需要 host 净化，避免 web 端口渗透到 MQTT/UDP 字段。
- `getSelectedConfig` 类的选择器一律要看 `Enabled`，`enabled=0` 就必须视作"不下发"。
- LLM 默认地址兜底只在用户未填 `base_url` 时生效，`type/base_url` 尽量在 seed 时写清楚，别依赖运行时兜底。
- 与"正常关闭"语义相关的日志（context canceled / EOF / DeadlineExceeded）该降级就降级，别把红字留给真实故障。
- 管理台自带的 OTA 测试设备（`ota-test-device`）短命 session 是常态，日志要分级。
- 未定案：后台 `handleDeviceActiveRequest` 兜底 `device_code` 匹配；OTA `external` 空/非空的策略；向导 host 净化实施。
