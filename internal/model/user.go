package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:50" json:"nickname"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Status    int            `gorm:"default:1" json:"status"` // 1:正常 2:禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserProfile 用户资料模型
type UserProfile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Gender    int       `gorm:"default:0" json:"gender"` // 0:未知 1:男 2:女
	Birthday  *time.Time `json:"birthday"`
	Bio       string    `gorm:"type:text" json:"bio"`
	Location  string    `gorm:"size:100" json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserStats 用户统计数据模型
type UserStats struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	GamesPlayed  int       `gorm:"default:0" json:"games_played"`
	GamesWon     int       `gorm:"default:0" json:"games_won"`
	GamesLost    int       `gorm:"default:0" json:"games_lost"`
	WinRate      float64   `gorm:"default:0" json:"win_rate"`
	TotalScore   int64     `gorm:"default:0" json:"total_score"`
	Level        int       `gorm:"default:1" json:"level"`
	Experience   int64     `gorm:"default:0" json:"experience"`
	LastPlayedAt *time.Time `json:"last_played_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 表名
func (UserStats) TableName() string {
	return "user_stats"
}

