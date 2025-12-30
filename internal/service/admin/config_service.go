package admin

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"github.com/iarna/toml"
	"github.com/game-apps/internal/utils"
)

// ConfigService 配置管理服务
type ConfigService struct {
	configBasePath string
}

// NewConfigService 创建配置管理服务
func NewConfigService(configBasePath string) *ConfigService {
	return &ConfigService{
		configBasePath: configBasePath,
	}
}

// GetConfig 获取服务配置
func (s *ConfigService) GetConfig(ctx context.Context, service string) (string, string, error) {
	var configPath string
	var fileType string

	switch service {
	case "backend":
		configPath = filepath.Join(s.configBasePath, "game-services", "configs", "config.yaml")
		fileType = "yaml"
	case "gateway":
		configPath = filepath.Join(s.configBasePath, "game-gateway", "config", "default.toml")
		fileType = "toml"
	case "agent":
		configPath = filepath.Join(s.configBasePath, "game-agent", "config", "config.yaml")
		fileType = "yaml"
	default:
		return "", "", utils.NewError(utils.ErrCodeInvalidInput, "不支持的服务类型")
	}

	// 如果配置文件不存在，尝试读取示例文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试读取示例文件
		examplePath := configPath + ".example"
		if _, err := os.Stat(examplePath); err == nil {
			configPath = examplePath
		} else {
			return "", "", utils.NewError(utils.ErrCodeNotFound, "配置文件不存在")
		}
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", "", utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("读取配置文件失败: %v", err))
	}

	return string(content), fileType, nil
}

// UpdateConfig 更新服务配置
func (s *ConfigService) UpdateConfig(ctx context.Context, service string, content string) error {
	var configPath string

	switch service {
	case "backend":
		configPath = filepath.Join(s.configBasePath, "game-services", "configs", "config.yaml")
	case "gateway":
		configPath = filepath.Join(s.configBasePath, "game-gateway", "config", "default.toml")
	case "agent":
		configPath = filepath.Join(s.configBasePath, "game-agent", "config", "config.yaml")
	default:
		return utils.NewError(utils.ErrCodeInvalidInput, "不支持的服务类型")
	}

	// 验证配置格式
	if err := s.ValidateConfig(service, content); err != nil {
		return err
	}

	// 创建备份
	backupPath := configPath + ".backup"
	if _, err := os.Stat(configPath); err == nil {
		originalContent, err := ioutil.ReadFile(configPath)
		if err == nil {
			ioutil.WriteFile(backupPath, originalContent, 0644)
		}
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("创建配置目录失败: %v", err))
	}

	// 写入新配置
	if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("写入配置文件失败: %v", err))
	}

	return nil
}

// ValidateConfig 验证配置格式
func (s *ConfigService) ValidateConfig(service string, content string) error {
	switch service {
	case "backend", "agent":
		var data interface{}
		if err := yaml.Unmarshal([]byte(content), &data); err != nil {
			return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("YAML 格式错误: %v", err))
		}
	case "gateway":
		var data interface{}
		if _, err := toml.Decode(content, &data); err != nil {
			return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("TOML 格式错误: %v", err))
		}
	default:
		return utils.NewError(utils.ErrCodeInvalidInput, "不支持的服务类型")
	}
	return nil
}

