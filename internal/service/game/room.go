package game

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/repository/redis"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

// RoomService 房间服务
type RoomService struct {
	roomRepo      RoomRepository
	roomPlayerRepo RoomPlayerRepository
	redisRoomRepo *redis.RoomRepository
	lockRepo      *redis.LockRepository
	logger        *zap.Logger
	maxPlayers     int
	defaultTimeout time.Duration
}

// RoomRepository 房间仓库接口
type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uint) (*model.Room, error)
	GetByRoomCode(ctx context.Context, roomCode string) (*model.Room, error)
	List(ctx context.Context, status *model.RoomStatus, limit, offset int) ([]*model.Room, error)
	Update(ctx context.Context, room *model.Room) error
	Delete(ctx context.Context, id uint) error
}

// RoomPlayerRepository 房间玩家仓库接口
type RoomPlayerRepository interface {
	Create(ctx context.Context, roomPlayer *model.RoomPlayer) error
	GetByRoomID(ctx context.Context, roomID uint) ([]*model.RoomPlayer, error)
	GetByRoomIDAndUserID(ctx context.Context, roomID, userID uint) (*model.RoomPlayer, error)
	Update(ctx context.Context, roomPlayer *model.RoomPlayer) error
	LeaveRoom(ctx context.Context, roomID, userID uint) error
}

// NewRoomService 创建房间服务
func NewRoomService(
	roomRepo RoomRepository,
	roomPlayerRepo RoomPlayerRepository,
	redisRoomRepo *redis.RoomRepository,
	lockRepo *redis.LockRepository,
	logger *zap.Logger,
	maxPlayers int,
	defaultTimeout time.Duration,
) *RoomService {
	return &RoomService{
		roomRepo:       roomRepo,
		roomPlayerRepo: roomPlayerRepo,
		redisRoomRepo:  redisRoomRepo,
		lockRepo:       lockRepo,
		logger:         logger,
		maxPlayers:     maxPlayers,
		defaultTimeout: defaultTimeout,
	}
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	Name     string `json:"name"`
	GameType string `json:"game_type"`
	Settings string `json:"settings"` // JSON 格式
}

// CreateRoomResponse 创建房间响应
type CreateRoomResponse struct {
	Room *model.Room `json:"room"`
}

// CreateRoom 创建房间
func (s *RoomService) CreateRoom(ctx context.Context, ownerID uint, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	// 生成房间代码
	roomCode, err := generateRoomCode()
	if err != nil {
		s.logger.Error("生成房间代码失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "创建房间失败")
	}

	// 设置过期时间
	expiresAt := time.Now().Add(s.defaultTimeout)

	// 创建房间
	room := &model.Room{
		RoomCode:       roomCode,
		Name:           req.Name,
		OwnerID:        ownerID,
		Status:         model.RoomStatusWaiting,
		MaxPlayers:     s.maxPlayers,
		CurrentPlayers: 0,
		GameType:       req.GameType,
		Settings:       req.Settings,
		ExpiresAt:      &expiresAt,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		s.logger.Error("创建房间失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "创建房间失败")
	}

	// 添加房主到房间
	roomPlayer := &model.RoomPlayer{
		RoomID:   room.ID,
		UserID:   ownerID,
		IsReady:  false,
		Position: 0,
		JoinedAt: time.Now(),
	}
	if err := s.roomPlayerRepo.Create(ctx, roomPlayer); err != nil {
		s.logger.Error("添加房主到房间失败", zap.Error(err))
		// 回滚：删除房间
		s.roomRepo.Delete(ctx, room.ID)
		return nil, utils.NewError(utils.ErrCodeInternal, "创建房间失败")
	}

	// 更新房间玩家数
	room.CurrentPlayers = 1
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("更新房间失败", zap.Error(err))
	}

	// 同步到 Redis
	s.syncRoomToRedis(ctx, room)

	return &CreateRoomResponse{
		Room: room,
	}, nil
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
}

// JoinRoomResponse 加入房间响应
type JoinRoomResponse struct {
	Room *model.Room `json:"room"`
}

// JoinRoom 加入房间
func (s *RoomService) JoinRoom(ctx context.Context, userID uint, req *JoinRoomRequest) (*JoinRoomResponse, error) {
	// 获取分布式锁
	lockKey := "room:lock:" + req.RoomCode
	acquired, err := s.lockRepo.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		s.logger.Error("获取锁失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "加入房间失败")
	}
	if !acquired {
		return nil, utils.NewError(utils.ErrCodeConflict, "房间正在被操作，请稍后重试")
	}
	defer s.lockRepo.ReleaseLock(ctx, lockKey)

	// 获取房间
	room, err := s.roomRepo.GetByRoomCode(ctx, req.RoomCode)
	if err != nil {
		s.logger.Error("查询房间失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "加入房间失败")
	}
	if room == nil {
		return nil, utils.NewError(utils.ErrCodeNotFound, "房间不存在")
	}

	// 检查房间状态
	if room.Status != model.RoomStatusWaiting {
		return nil, utils.NewError(utils.ErrCodeConflict, "房间已开始或已结束")
	}

	// 检查房间是否已满
	if room.CurrentPlayers >= room.MaxPlayers {
		return nil, utils.NewError(utils.ErrCodeConflict, "房间已满")
	}

	// 检查是否已在房间中
	existingPlayer, err := s.roomPlayerRepo.GetByRoomIDAndUserID(ctx, room.ID, userID)
	if err != nil {
		s.logger.Error("查询房间玩家失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "加入房间失败")
	}
	if existingPlayer != nil {
		return nil, utils.NewError(utils.ErrCodeConflict, "已在房间中")
	}

	// 添加玩家到房间
	players, err := s.roomPlayerRepo.GetByRoomID(ctx, room.ID)
	if err != nil {
		s.logger.Error("查询房间玩家失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "加入房间失败")
	}

	roomPlayer := &model.RoomPlayer{
		RoomID:   room.ID,
		UserID:   userID,
		IsReady:  false,
		Position: len(players),
		JoinedAt: time.Now(),
	}
	if err := s.roomPlayerRepo.Create(ctx, roomPlayer); err != nil {
		s.logger.Error("添加玩家到房间失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "加入房间失败")
	}

	// 更新房间玩家数
	room.CurrentPlayers++
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("更新房间失败", zap.Error(err))
	}

	// 同步到 Redis
	s.syncRoomToRedis(ctx, room)
	s.redisRoomRepo.AddRoomPlayer(ctx, room.ID, userID)

	return &JoinRoomResponse{
		Room: room,
	}, nil
}

// LeaveRoom 离开房间
func (s *RoomService) LeaveRoom(ctx context.Context, userID uint, roomID uint) error {
	// 获取分布式锁
	lockKey := "room:lock:" + string(rune(roomID))
	acquired, err := s.lockRepo.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		s.logger.Error("获取锁失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "离开房间失败")
	}
	if !acquired {
		return utils.NewError(utils.ErrCodeConflict, "房间正在被操作，请稍后重试")
	}
	defer s.lockRepo.ReleaseLock(ctx, lockKey)

	// 获取房间
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Error("查询房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "离开房间失败")
	}
	if room == nil {
		return utils.NewError(utils.ErrCodeNotFound, "房间不存在")
	}

	// 离开房间
	if err := s.roomPlayerRepo.LeaveRoom(ctx, roomID, userID); err != nil {
		s.logger.Error("离开房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "离开房间失败")
	}

	// 更新房间玩家数
	if room.CurrentPlayers > 0 {
		room.CurrentPlayers--
		if err := s.roomRepo.Update(ctx, room); err != nil {
			s.logger.Error("更新房间失败", zap.Error(err))
		}
	}

	// 如果房间为空，删除房间
	if room.CurrentPlayers == 0 {
		if err := s.roomRepo.Delete(ctx, roomID); err != nil {
			s.logger.Error("删除房间失败", zap.Error(err))
		}
		s.redisRoomRepo.DeleteRoom(ctx, roomID)
	} else {
		// 同步到 Redis
		s.syncRoomToRedis(ctx, room)
		s.redisRoomRepo.RemoveRoomPlayer(ctx, roomID, userID)
	}

	return nil
}

// GetRoom 获取房间信息
func (s *RoomService) GetRoom(ctx context.Context, roomID uint) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Error("查询房间失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "获取房间失败")
	}
	if room == nil {
		return nil, utils.NewError(utils.ErrCodeNotFound, "房间不存在")
	}
	return room, nil
}

// ListRooms 列出房间
func (s *RoomService) ListRooms(ctx context.Context, status *model.RoomStatus, limit, offset int) ([]*model.Room, error) {
	return s.roomRepo.List(ctx, status, limit, offset)
}

// syncRoomToRedis 同步房间到 Redis
func (s *RoomService) syncRoomToRedis(ctx context.Context, room *model.Room) {
	roomData := map[string]interface{}{
		"id":              room.ID,
		"room_code":       room.RoomCode,
		"name":            room.Name,
		"owner_id":        room.OwnerID,
		"status":          room.Status,
		"max_players":     room.MaxPlayers,
		"current_players": room.CurrentPlayers,
		"game_type":      room.GameType,
		"settings":        room.Settings,
	}
	if room.ExpiresAt != nil {
		roomData["expires_at"] = room.ExpiresAt.Unix()
	}
	s.redisRoomRepo.SetRoomState(ctx, room.ID, roomData, s.defaultTimeout)
}

// generateRoomCode 生成房间代码
func generateRoomCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

