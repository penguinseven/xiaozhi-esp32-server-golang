package user_config

import (
	"fmt"
	"sync"

	"xiaozhi-esp32-server-golang/internal/domain/config/manager"
	"xiaozhi-esp32-server-golang/internal/domain/config/memory"
	userconfig_redis "xiaozhi-esp32-server-golang/internal/domain/config/redis"
	"xiaozhi-esp32-server-golang/internal/util"
)

// Config 用户配置提供者配置结构
type Config struct {
	Type       string                 `json:"type"`       // 存储类型: "redis", "memory", "file"
	Parameters map[string]interface{} `json:"parameters"` // 存储相关配置参数
}

var (
	providerOnce     sync.Once
	providerInstance UserConfigProvider
	providerErr      error
)

// GetProvider 返回单例 ConfigProvider。首次调用时构造，后续调用返回同一实例。
// sType 仅首次调用生效；调用方应传入 viper.GetString("config_provider.type")。
func GetProvider(sType string) (UserConfigProvider, error) {
	providerOnce.Do(func() {
		config := make(map[string]interface{})
		if sType == "manager" {
			backendUrl := util.GetBackendURL()
			config = map[string]interface{}{
				"backend_url": backendUrl,
				"auth_token":  util.GetManagerAuthToken(),
			}
		}
		providerInstance, providerErr = GetUserConfigProvider(sType, config)
	})
	return providerInstance, providerErr
}

// GetUserConfigProvider 创建用户配置提供者（不缓存，每次构造新实例）
func GetUserConfigProvider(providerType string, config map[string]interface{}) (UserConfigProvider, error) {
	if config == nil {
		config = make(map[string]interface{})
	}

	switch providerType {
	case "redis":
		provider, err := userconfig_redis.NewRedisUserConfigProvider(config)
		if err != nil {
			return nil, fmt.Errorf("创建Redis用户配置提供者失败: %v", err)
		}
		return provider, nil
	case "manager":
		provider, err := manager.NewManagerUserConfigProvider(config)
		if err != nil {
			return nil, fmt.Errorf("创建后端管理系统用户配置提供者失败: %v", err)
		}
		return provider, nil
	case "memory":
		provider, err := memory.NewMemoryUserConfigProvider(config)
		if err != nil {
			return nil, fmt.Errorf("创建内存用户配置提供者失败: %v", err)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("不支持的用户配置提供者: %s", providerType)
	}
}
