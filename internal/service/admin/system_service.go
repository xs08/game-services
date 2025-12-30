package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/game-apps/internal/utils"
)

// SystemService 系统配置管理服务
type SystemService struct {
	configPath string
}

// NewSystemService 创建系统配置管理服务
func NewSystemService(configBasePath string) *SystemService {
	configPath := filepath.Join(configBasePath, "game-services", "configs", "system_config.json")
	return &SystemService{
		configPath: configPath,
	}
}

// SystemConfig 系统配置结构
type SystemConfig struct {
	Basic        BasicConfig        `json:"basic"`
	Security     SecurityConfig     `json:"security"`
	Notification NotificationConfig `json:"notification"`
}

type BasicConfig struct {
	Timezone        string `json:"timezone"`
	Language        string `json:"language"`
	Theme           string `json:"theme"`
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	SiteLogo        string `json:"site_logo"`
}

type SecurityConfig struct {
	PasswordPolicy PasswordPolicy `json:"password_policy"`
	IPWhitelist    []string       `json:"ip_whitelist"`
	JWT            JWTConfig      `json:"jwt"`
	Session        SessionConfig  `json:"session"`
}

type PasswordPolicy struct {
	MinLength          int  `json:"min_length"`
	RequireUppercase   bool `json:"require_uppercase"`
	RequireLowercase   bool `json:"require_lowercase"`
	RequireNumbers     bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	ExpirationDays     int  `json:"expiration_days"`
}

type JWTConfig struct {
	Secret                string `json:"secret"`
	ExpirationHours       int    `json:"expiration_hours"`
	RefreshExpirationHours int    `json:"refresh_expiration_hours"`
}

type SessionConfig struct {
	TimeoutMinutes        int `json:"timeout_minutes"`
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
}

type NotificationConfig struct {
	Email EmailConfig `json:"email"`
	SMS   SMSConfig   `json:"sms"`
	Push  PushConfig  `json:"push"`
}

type EmailConfig struct {
	Enabled    bool   `json:"enabled"`
	SMTPHost   string `json:"smtp_host"`
	SMTPPort   int    `json:"smtp_port"`
	SMTPUser   string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
}

type SMSConfig struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type PushConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

// GetSystemConfig 获取系统配置
func (s *SystemService) GetSystemConfig(ctx context.Context) (*SystemConfig, error) {
	var config SystemConfig

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return s.getDefaultConfig(), nil
	}

	content, err := ioutil.ReadFile(s.configPath)
	if err != nil {
		return nil, utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("读取系统配置文件失败: %v", err))
	}

	if err := json.Unmarshal(content, &config); err != nil {
		return nil, utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("解析系统配置文件失败: %v", err))
	}

	return &config, nil
}

// GetSystemConfigCategory 获取分类配置
func (s *SystemService) GetSystemConfigCategory(ctx context.Context, category string) (interface{}, error) {
	config, err := s.GetSystemConfig(ctx)
	if err != nil {
		return nil, err
	}

	switch category {
	case "basic":
		return config.Basic, nil
	case "security":
		return config.Security, nil
	case "notification":
		return config.Notification, nil
	default:
		return nil, utils.NewError(utils.ErrCodeInvalidInput, "不支持的配置分类")
	}
}

// UpdateSystemConfig 更新系统配置
func (s *SystemService) UpdateSystemConfig(ctx context.Context, updates *SystemConfig) error {
	config, err := s.GetSystemConfig(ctx)
	if err != nil {
		return err
	}

	// 合并更新
	if updates.Basic.SiteName != "" {
		config.Basic = updates.Basic
	}
	if len(updates.Security.IPWhitelist) > 0 || updates.Security.PasswordPolicy.MinLength > 0 {
		config.Security = updates.Security
	}
	if updates.Notification.Email.SMTPHost != "" || updates.Notification.SMS.Provider != "" || updates.Notification.Push.Provider != "" {
		config.Notification = updates.Notification
	}

	// 保存配置
	return s.saveConfig(config)
}

// UpdateSystemConfigCategory 更新分类配置
func (s *SystemService) UpdateSystemConfigCategory(ctx context.Context, category string, data interface{}) error {
	config, err := s.GetSystemConfig(ctx)
	if err != nil {
		return err
	}

	// 将数据转换为 JSON 再解析，以支持部分更新
	jsonData, err := json.Marshal(data)
	if err != nil {
		return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("数据格式错误: %v", err))
	}

	switch category {
	case "basic":
		if err := json.Unmarshal(jsonData, &config.Basic); err != nil {
			return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("基础配置格式错误: %v", err))
		}
	case "security":
		if err := json.Unmarshal(jsonData, &config.Security); err != nil {
			return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("安全配置格式错误: %v", err))
		}
	case "notification":
		if err := json.Unmarshal(jsonData, &config.Notification); err != nil {
			return utils.NewError(utils.ErrCodeInvalidInput, fmt.Sprintf("通知配置格式错误: %v", err))
		}
	default:
		return utils.NewError(utils.ErrCodeInvalidInput, "不支持的配置分类")
	}

	return s.saveConfig(config)
}

func (s *SystemService) saveConfig(config *SystemConfig) error {
	// 创建备份
	backupPath := s.configPath + ".backup"
	if _, err := os.Stat(s.configPath); err == nil {
		originalContent, err := ioutil.ReadFile(s.configPath)
		if err == nil {
			ioutil.WriteFile(backupPath, originalContent, 0644)
		}
	}

	// 确保目录存在
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("创建配置目录失败: %v", err))
	}

	// 写入配置
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("序列化配置失败: %v", err))
	}

	if err := ioutil.WriteFile(s.configPath, jsonData, 0644); err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("写入配置文件失败: %v", err))
	}

	return nil
}

func (s *SystemService) getDefaultConfig() *SystemConfig {
	return &SystemConfig{
		Basic: BasicConfig{
			Timezone:        "Asia/Shanghai",
			Language:        "zh-CN",
			Theme:           "light",
			SiteName:        "游戏服务管理控制台",
			SiteDescription: "",
			SiteLogo:        "",
		},
		Security: SecurityConfig{
			PasswordPolicy: PasswordPolicy{
				MinLength:          8,
				RequireUppercase:   true,
				RequireLowercase:   true,
				RequireNumbers:     true,
				RequireSpecialChars: false,
				ExpirationDays:     90,
			},
			IPWhitelist: []string{},
			JWT: JWTConfig{
				Secret:                 "",
				ExpirationHours:        24,
				RefreshExpirationHours: 168,
			},
			Session: SessionConfig{
				TimeoutMinutes:        30,
				MaxConcurrentSessions: 5,
			},
		},
		Notification: NotificationConfig{
			Email: EmailConfig{
				Enabled:    false,
				SMTPHost:   "",
				SMTPPort:   587,
				SMTPUser:   "",
				SMTPPassword: "",
				FromEmail:  "",
				FromName:   "",
			},
			SMS: SMSConfig{
				Enabled:  false,
				Provider: "",
				APIKey:   "",
				APISecret: "",
			},
			Push: PushConfig{
				Enabled:  false,
				Provider: "",
				APIKey:   "",
			},
		},
	}
}

