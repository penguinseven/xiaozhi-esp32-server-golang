package constants

// ConfigProviderType 配置来源类型常量
const (
	ConfigProviderTypeMemory  = "memory"
	ConfigProviderTypeManager = "manager"
	ConfigProviderTypeRedis   = "redis"
)

// ConfigSource 常量（词汇对齐 CONTEXT.md）
const (
	ConfigSourceDefaults  = "defaults"
	ConfigSourceBase      = "base"
	ConfigSourceLocal     = "local"
	ConfigSourceEnv       = "env"
	ConfigSourceRemote    = "remote"
	ConfigSourceHotReload = "hot_reload"
)

// ResolvePhase 解析阶段
const (
	ResolvePhaseDefaults = iota
	ResolvePhaseBase
	ResolvePhaseLocal
	ResolvePhaseEnv
	ResolvePhaseRemote
	ResolvePhaseHotReload
)

// 敏感键集合 —— 这些键的 env 来源优先级高于 remote（避免远程覆盖本地密钥）
var SensitiveKeys = map[string]bool{
	"password":            true,
	"api_key":             true,
	"api_secret":          true,
	"access_token":        true,
	"auth_token":          true,
	"endpoint_auth_token": true,
	"secret":              true,
	"token":               true,
	"private_key":         true,
}

// 允许从环境变量绑定的键白名单（XIAOZHI_ 前缀）
var BindEnvAllowlist = map[string]bool{
	"config_provider.type":             true,
	"config_provider.manager.url":      true,
	"config_provider.manager.username": true,
	"config_provider.manager.password": true,
	"config_provider.redis.addr":       true,
	"config_provider.redis.password":   true,
	"config_provider.redis.db":         true,
	"mqtt.broker":                      true,
	"mqtt.port":                        true,
	"mqtt.username":                    true,
	"mqtt.password":                    true,
	"server.port":                      true,
	"server.host":                      true,
	"log.level":                        true,
	"redis.addr":                       true,
	"redis.password":                   true,
	"redis.db":                         true,
	"asr.type":                         true,
	"llm.type":                         true,
	"tts.type":                         true,
	"vad.type":                         true,
}
