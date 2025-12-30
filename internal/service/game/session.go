package game

import (
	"context"
	"time"

	"github.com/game-apps/internal/repository/redis"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

// SessionService 会话服务
type SessionService struct {
	sessionRepo    *redis.SessionRepository
	onlineUserRepo *redis.OnlineUserRepository
	logger         *zap.Logger
	heartbeatInterval time.Duration
	timeout          time.Duration
}

// NewSessionService 创建会话服务
func NewSessionService(
	sessionRepo *redis.SessionRepository,
	onlineUserRepo *redis.OnlineUserRepository,
	logger *zap.Logger,
	heartbeatInterval, timeout time.Duration,
) *SessionService {
	return &SessionService{
		sessionRepo:       sessionRepo,
		onlineUserRepo:    onlineUserRepo,
		logger:            logger,
		heartbeatInterval: heartbeatInterval,
		timeout:           timeout,
	}
}

// CreateSession 创建会话
func (s *SessionService) CreateSession(ctx context.Context, userID uint, ipAddress, userAgent string) error {
	// 保存会话信息
	sessionData := map[string]interface{}{
		"user_id":       userID,
		"ip_address":    ipAddress,
		"user_agent":    userAgent,
		"last_activity": time.Now().Unix(),
		"status":        1, // 在线
	}

	if err := s.sessionRepo.SetSession(ctx, userID, sessionData, s.timeout); err != nil {
		s.logger.Error("保存会话失败", zap.Error(err), zap.Uint("user_id", userID))
		return utils.NewError(utils.ErrCodeInternal, "创建会话失败")
	}

	// 添加到在线用户列表
	if err := s.onlineUserRepo.AddOnlineUser(ctx, userID); err != nil {
		s.logger.Warn("添加在线用户失败", zap.Error(err))
	}

	return nil
}

// UpdateSessionActivity 更新会话活动时间
func (s *SessionService) UpdateSessionActivity(ctx context.Context, userID uint) error {
	sessionData, err := s.sessionRepo.GetSession(ctx, userID)
	if err != nil {
		// 会话不存在，创建新会话
		return s.CreateSession(ctx, userID, "", "")
	}

	sessionData["last_activity"] = time.Now().Unix()
	if err := s.sessionRepo.SetSession(ctx, userID, sessionData, s.timeout); err != nil {
		s.logger.Error("更新会话失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "更新会话失败")
	}

	return nil
}

// GetSession 获取会话
func (s *SessionService) GetSession(ctx context.Context, userID uint) (map[string]interface{}, error) {
	return s.sessionRepo.GetSession(ctx, userID)
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(ctx context.Context, userID uint) error {
	// 删除会话
	if err := s.sessionRepo.DeleteSession(ctx, userID); err != nil {
		s.logger.Error("删除会话失败", zap.Error(err))
	}

	// 从在线用户列表移除
	if err := s.onlineUserRepo.RemoveOnlineUser(ctx, userID); err != nil {
		s.logger.Warn("移除在线用户失败", zap.Error(err))
	}

	return nil
}

// IsOnline 检查用户是否在线
func (s *SessionService) IsOnline(ctx context.Context, userID uint) (bool, error) {
	return s.onlineUserRepo.IsOnline(ctx, userID)
}

// GetOnlineUsers 获取所有在线用户
func (s *SessionService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return s.onlineUserRepo.GetOnlineUsers(ctx)
}

// CheckSessionTimeout 检查会话超时
func (s *SessionService) CheckSessionTimeout(ctx context.Context, userID uint) (bool, error) {
	sessionData, err := s.sessionRepo.GetSession(ctx, userID)
	if err != nil {
		return true, nil // 会话不存在，视为超时
	}

	lastActivity, ok := sessionData["last_activity"].(float64)
	if !ok {
		return true, nil
	}

	lastActivityTime := time.Unix(int64(lastActivity), 0)
	timeoutTime := lastActivityTime.Add(s.timeout)

	return time.Now().After(timeoutTime), nil
}

