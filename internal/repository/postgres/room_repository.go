package postgres

import (
	"context"
	"errors"

	"github.com/game-apps/internal/model"
	"gorm.io/gorm"
)

// RoomRepository 房间数据访问层（PostgreSQL）
type RoomRepository struct {
	db *gorm.DB
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create 创建房间
func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

// GetByID 根据 ID 获取房间
func (r *RoomRepository) GetByID(ctx context.Context, id uint) (*model.Room, error) {
	var room model.Room
	err := r.db.WithContext(ctx).First(&room, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// GetByRoomCode 根据房间代码获取房间
func (r *RoomRepository) GetByRoomCode(ctx context.Context, roomCode string) (*model.Room, error) {
	var room model.Room
	err := r.db.WithContext(ctx).Where("room_code = ?", roomCode).First(&room).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// List 列出房间
func (r *RoomRepository) List(ctx context.Context, status *model.RoomStatus, limit, offset int) ([]*model.Room, error) {
	var rooms []*model.Room
	query := r.db.WithContext(ctx)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rooms).Error
	return rooms, err
}

// Update 更新房间
func (r *RoomRepository) Update(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Save(room).Error
}

// Delete 删除房间（软删除）
func (r *RoomRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Room{}, id).Error
}

// RoomPlayerRepository 房间玩家数据访问层
type RoomPlayerRepository struct {
	db *gorm.DB
}

// NewRoomPlayerRepository 创建房间玩家仓库
func NewRoomPlayerRepository(db *gorm.DB) *RoomPlayerRepository {
	return &RoomPlayerRepository{db: db}
}

// Create 创建房间玩家关系
func (r *RoomPlayerRepository) Create(ctx context.Context, roomPlayer *model.RoomPlayer) error {
	return r.db.WithContext(ctx).Create(roomPlayer).Error
}

// GetByRoomID 根据房间 ID 获取所有玩家
func (r *RoomPlayerRepository) GetByRoomID(ctx context.Context, roomID uint) ([]*model.RoomPlayer, error) {
	var players []*model.RoomPlayer
	err := r.db.WithContext(ctx).Where("room_id = ? AND left_at IS NULL", roomID).Find(&players).Error
	return players, err
}

// GetByRoomIDAndUserID 根据房间 ID 和用户 ID 获取关系
func (r *RoomPlayerRepository) GetByRoomIDAndUserID(ctx context.Context, roomID, userID uint) (*model.RoomPlayer, error) {
	var player model.RoomPlayer
	err := r.db.WithContext(ctx).Where("room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

// Update 更新房间玩家关系
func (r *RoomPlayerRepository) Update(ctx context.Context, roomPlayer *model.RoomPlayer) error {
	return r.db.WithContext(ctx).Save(roomPlayer).Error
}

// LeaveRoom 离开房间
func (r *RoomPlayerRepository) LeaveRoom(ctx context.Context, roomID, userID uint) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).
		Model(&model.RoomPlayer{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("left_at", now).Error
}

