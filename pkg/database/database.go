package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Driver          string
	MySQLConfig     MySQLConfig
	PostgresConfig  PostgresConfig
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	Charset   string
	ParseTime bool
	Loc       string
}

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect 连接数据库
func Connect(config Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			config.MySQLConfig.User,
			config.MySQLConfig.Password,
			config.MySQLConfig.Host,
			config.MySQLConfig.Port,
			config.MySQLConfig.DBName,
			config.MySQLConfig.Charset,
			config.MySQLConfig.ParseTime,
			config.MySQLConfig.Loc,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.PostgresConfig.Host,
			config.PostgresConfig.User,
			config.PostgresConfig.Password,
			config.PostgresConfig.DBName,
			config.PostgresConfig.Port,
			config.PostgresConfig.SSLMode,
		)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}

