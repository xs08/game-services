package game

import (
	"context"
	"encoding/json"
	"time"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/repository/redis"
	"github.com/game-apps/internal/utils"
	"github.com/game-apps/pkg/cache"
	"go.uber.org/zap"
)

// GameState 游戏状态
type GameState int

const (
	GameStateWaiting   GameState = 1 // 等待中
	GameStateStarting  GameState = 2 // 开始中
	GameStatePlaying   GameState = 3 // 进行中
	GameStatePaused    GameState = 4 // 暂停
	GameStateFinished  GameState = 5 // 已结束
)

// GameEvent 游戏事件
type GameEvent struct {
	Type      string                 `json:"type"`
	RoomID    uint                   `json:"room_id"`
	UserID    uint                   `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// ProcessService 游戏逻辑进程服务
type ProcessService struct {
	roomRepo      RoomRepository
	redisRoomRepo *redis.RoomRepository
	lockRepo      *redis.LockRepository
	logger        *zap.Logger
	eventChannel  string
}

// NewProcessService 创建游戏进程服务
func NewProcessService(
	roomRepo RoomRepository,
	redisRoomRepo *redis.RoomRepository,
	lockRepo *redis.LockRepository,
	logger *zap.Logger,
	eventChannel string,
) *ProcessService {
	cacheClient := redisRoomRepo.Client()
	return &ProcessService{
		roomRepo:      roomRepo,
		redisRoomRepo: redisRoomRepo,
		lockRepo:      lockRepo,
		logger:        logger,
		eventChannel:  eventChannel,
		cacheClient:   cacheClient,
	}
}

// StartGame 开始游戏
func (s *ProcessService) StartGame(ctx context.Context, roomID uint) error {
	// 获取分布式锁
	lockKey := "game:lock:" + string(rune(roomID))
	acquired, err := s.lockRepo.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil {
		s.logger.Error("获取锁失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "开始游戏失败")
	}
	if !acquired {
		return utils.NewError(utils.ErrCodeConflict, "游戏正在被操作，请稍后重试")
	}
	defer s.lockRepo.ReleaseLock(ctx, lockKey)

	// 获取房间
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Error("查询房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "开始游戏失败")
	}
	if room == nil {
		return utils.NewError(utils.ErrCodeNotFound, "房间不存在")
	}

	// 检查房间状态
	if room.Status != model.RoomStatusWaiting {
		return utils.NewError(utils.ErrCodeConflict, "房间状态不允许开始游戏")
	}

	// 更新房间状态
	now := time.Now()
	room.Status = model.RoomStatusPlaying
	room.StartedAt = &now
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("更新房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "开始游戏失败")
	}

	// 同步到 Redis
	roomData := map[string]interface{}{
		"status":    room.Status,
		"started_at": now.Unix(),
		"game_state": GameStateStarting,
	}
	s.redisRoomRepo.SetRoomState(ctx, roomID, roomData, 0)

	// 发布游戏开始事件
	event := &GameEvent{
		Type:      "game_start",
		RoomID:    roomID,
		Data:      map[string]interface{}{"room": room},
		Timestamp: time.Now().Unix(),
	}
	if err := s.PublishEvent(ctx, event); err != nil {
		s.logger.Warn("发布事件失败", zap.Error(err))
	}

	return nil
}

// EndGame 结束游戏
func (s *ProcessService) EndGame(ctx context.Context, roomID uint, results map[uint]interface{}) error {
	// 获取分布式锁
	lockKey := "game:lock:" + string(rune(roomID))
	acquired, err := s.lockRepo.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil {
		s.logger.Error("获取锁失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "结束游戏失败")
	}
	if !acquired {
		return utils.NewError(utils.ErrCodeConflict, "游戏正在被操作，请稍后重试")
	}
	defer s.lockRepo.ReleaseLock(ctx, lockKey)

	// 获取房间
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Error("查询房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "结束游戏失败")
	}
	if room == nil {
		return utils.NewError(utils.ErrCodeNotFound, "房间不存在")
	}

	// 更新房间状态
	now := time.Now()
	room.Status = model.RoomStatusFinished
	room.EndedAt = &now
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("更新房间失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "结束游戏失败")
	}

	// 同步到 Redis
	roomData := map[string]interface{}{
		"status":    room.Status,
		"ended_at":  now.Unix(),
		"game_state": GameStateFinished,
		"results":    results,
	}
	s.redisRoomRepo.SetRoomState(ctx, roomID, roomData, 0)

	// 发布游戏结束事件
	event := &GameEvent{
		Type:      "game_end",
		RoomID:    roomID,
		Data:      map[string]interface{}{"room": room, "results": results},
		Timestamp: time.Now().Unix(),
	}
	if err := s.PublishEvent(ctx, event); err != nil {
		s.logger.Warn("发布事件失败", zap.Error(err))
	}

	return nil
}

// UpdateGameState 更新游戏状态
func (s *ProcessService) UpdateGameState(ctx context.Context, roomID uint, state GameState, data map[string]interface{}) error {
	roomData := map[string]interface{}{
		"game_state": state,
	}
	for k, v := range data {
		roomData[k] = v
	}
	return s.redisRoomRepo.SetRoomState(ctx, roomID, roomData, 0)
}

// GetGameState 获取游戏状态
func (s *ProcessService) GetGameState(ctx context.Context, roomID uint) (map[string]string, error) {
	return s.redisRoomRepo.GetRoomState(ctx, roomID)
}

// PublishEvent 发布游戏事件
func (s *ProcessService) PublishEvent(ctx context.Context, event *GameEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 使用 Redis Pub/Sub 发布事件
	if s.cacheClient == nil {
		return utils.NewError(utils.ErrCodeInternal, "Redis 客户端不可用")
	}

	return s.cacheClient.GetClient().Publish(ctx, s.eventChannel, eventData).Err()
}

// SubscribeEvents 订阅游戏事件
func (s *ProcessService) SubscribeEvents(ctx context.Context) (<-chan *GameEvent, error) {
	if s.cacheClient == nil {
		return nil, utils.NewError(utils.ErrCodeInternal, "Redis 客户端不可用")
	}

	pubsub := s.cacheClient.GetClient().Subscribe(ctx, s.eventChannel)
	eventChan := make(chan *GameEvent, 100)

	go func() {
		defer close(eventChan)
		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				s.logger.Error("接收消息失败", zap.Error(err))
				return
			}

			var event GameEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				s.logger.Error("解析事件失败", zap.Error(err))
				continue
			}

			select {
			case eventChan <- &event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, nil
}


