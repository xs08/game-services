package model

import (
	"time"

	"gorm.io/gorm"
)

// RoomStatus 房间状态
type RoomStatus int

const (
	RoomStatusWaiting   RoomStatus = 1 // 等待中
	RoomStatusPlaying   RoomStatus = 2 // 进行中
	RoomStatusFinished  RoomStatus = 3 // 已结束
	RoomStatusCancelled RoomStatus = 4 // 已取消
)

// Room 房间模型
type Room struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RoomCode    string         `gorm:"uniqueIndex;size:20;not null" json:"room_code"`
	Name        string         `gorm:"size:100" json:"name"`
	OwnerID     uint           `gorm:"not null" json:"owner_id"`
	Status      RoomStatus     `gorm:"default:1" json:"status"`
	MaxPlayers  int            `gorm:"default:10" json:"max_players"`
	CurrentPlayers int         `gorm:"default:0" json:"current_players"`
	GameType    string         `gorm:"size:50" json:"game_type"`
	Settings    string         `gorm:"type:text" json:"settings"` // JSON 格式的游戏设置
	StartedAt   *time.Time     `json:"started_at"`
	EndedAt     *time.Time     `json:"ended_at"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Room) TableName() string {
	return "rooms"
}

// RoomPlayer 房间玩家关系模型
type RoomPlayer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RoomID     uint      `gorm:"index;not null" json:"room_id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	IsReady    bool      `gorm:"default:false" json:"is_ready"`
	Position   int       `gorm:"default:0" json:"position"` // 在房间中的位置
	JoinedAt   time.Time `json:"joined_at"`
	LeftAt     *time.Time `json:"left_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 表名
func (RoomPlayer) TableName() string {
	return "room_players"
}

