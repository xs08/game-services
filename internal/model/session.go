package model

import (
	"time"

	"gorm.io/gorm"
)

// SessionStatus 会话状态
type SessionStatus int

const (
	SessionStatusOnline  SessionStatus = 1 // 在线
	SessionStatusOffline SessionStatus = 2 // 离线
	SessionStatusAway   SessionStatus = 3 // 离开
)

// Session 会话模型（主要用于审计）
type Session struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	Token        string         `gorm:"index;size:255;not null" json:"-"`
	IPAddress    string         `gorm:"size:45" json:"ip_address"`
	UserAgent    string         `gorm:"size:255" json:"user_agent"`
	Status       SessionStatus  `gorm:"default:1" json:"status"`
	LastActivity time.Time      `json:"last_activity"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Session) TableName() string {
	return "sessions"
}

