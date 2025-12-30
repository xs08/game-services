package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Log        LogConfig        `mapstructure:"log"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Game       GameConfig        `mapstructure:"game"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	HTTPPort     int           `mapstructure:"http_port"`
	GRPCPort     int           `mapstructure:"grpc_port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Driver   string         `mapstructure:"driver"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type JWTConfig struct {
	Secret                string `mapstructure:"secret"`
	ExpirationHours       int    `mapstructure:"expiration_hours"`
	RefreshExpirationHours int    `mapstructure:"refresh_expiration_hours"`
}

type LogConfig struct {
	Level  string     `mapstructure:"level"`
	Format string     `mapstructure:"format"`
	Output string     `mapstructure:"output"`
	File   LogFileConfig `mapstructure:"file"`
}

type LogFileConfig struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type MonitoringConfig struct {
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	MetricsPath    string `mapstructure:"metrics_path"`
	HealthPath     string `mapstructure:"health_path"`
	ReadyPath      string `mapstructure:"ready_path"`
}

type GameConfig struct {
	Room    RoomConfig    `mapstructure:"room"`
	Session SessionConfig `mapstructure:"session"`
}

type RoomConfig struct {
	MaxPlayers     int           `mapstructure:"max_players"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type SessionConfig struct {
	HeartbeatInterval  time.Duration `mapstructure:"heartbeat_interval"`
	Timeout            time.Duration `mapstructure:"timeout"`
	MaxReconnectAttempts int         `mapstructure:"max_reconnect_attempts"`
}

var globalConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")
	}

	// 环境变量支持
	viper.SetEnvPrefix("GAME_APPS")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("HTTP 端口无效: %d", c.Server.HTTPPort)
	}

	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("gRPC 端口无效: %d", c.Server.GRPCPort)
	}

	if c.Database.Driver != "mysql" && c.Database.Driver != "postgres" {
		return fmt.Errorf("不支持的数据库驱动: %s", c.Database.Driver)
	}

	if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
		return fmt.Errorf("JWT secret 未设置或使用默认值")
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.grpc_port", 9090)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")

	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.mysql.host", "localhost")
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.max_open_conns", 100)
	viper.SetDefault("database.mysql.max_idle_conns", 10)

	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.sslmode", "disable")
	viper.SetDefault("database.postgres.max_open_conns", 100)
	viper.SetDefault("database.postgres.max_idle_conns", 10)

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 100)
	viper.SetDefault("redis.min_idle_conns", 10)

	viper.SetDefault("jwt.expiration_hours", 24)
	viper.SetDefault("jwt.refresh_expiration_hours", 168)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	viper.SetDefault("monitoring.metrics_enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_path", "/health")
	viper.SetDefault("monitoring.ready_path", "/ready")

	viper.SetDefault("game.room.max_players", 10)
	viper.SetDefault("game.room.default_timeout", "300s")
	viper.SetDefault("game.session.heartbeat_interval", "30s")
	viper.SetDefault("game.session.timeout", "120s")
}

