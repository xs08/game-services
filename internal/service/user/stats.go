package user

import (
	"context"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

// StatsService 用户统计服务
type StatsService struct {
	userStatsRepo UserStatsRepository
	logger        *zap.Logger
}

// NewStatsService 创建用户统计服务
func NewStatsService(
	userStatsRepo UserStatsRepository,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		userStatsRepo: userStatsRepo,
		logger:        logger,
	}
}

// GetStatsRequest 获取统计请求
type GetStatsRequest struct {
	UserID uint
}

// GetStatsResponse 获取统计响应
type GetStatsResponse struct {
	Stats *model.UserStats `json:"stats"`
}

// GetStats 获取用户统计
func (s *StatsService) GetStats(ctx context.Context, userID uint) (*GetStatsResponse, error) {
	stats, err := s.userStatsRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户统计失败", zap.Error(err), zap.Uint("user_id", userID))
		return nil, utils.NewError(utils.ErrCodeInternal, "获取统计失败")
	}

	if stats == nil {
		// 创建默认统计
		stats = &model.UserStats{
			UserID: userID,
		}
		if err := s.userStatsRepo.Create(ctx, stats); err != nil {
			s.logger.Error("创建用户统计失败", zap.Error(err))
			return nil, utils.NewError(utils.ErrCodeInternal, "获取统计失败")
		}
	}

	return &GetStatsResponse{
		Stats: stats,
	}, nil
}

// UpdateGameResult 更新游戏结果
func (s *StatsService) UpdateGameResult(ctx context.Context, userID uint, won bool, score int64) error {
	stats, err := s.userStatsRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户统计失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "更新统计失败")
	}

	if stats == nil {
		stats = &model.UserStats{
			UserID: userID,
		}
	}

	stats.GamesPlayed++
	if won {
		stats.GamesWon++
	} else {
		stats.GamesLost++
	}
	stats.TotalScore += score

	// 更新胜率
	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
	}

	if stats.ID == 0 {
		if err := s.userStatsRepo.Create(ctx, stats); err != nil {
			s.logger.Error("创建用户统计失败", zap.Error(err))
			return utils.NewError(utils.ErrCodeInternal, "更新统计失败")
		}
	} else {
		if err := s.userStatsRepo.Update(ctx, stats); err != nil {
			s.logger.Error("更新用户统计失败", zap.Error(err))
			return utils.NewError(utils.ErrCodeInternal, "更新统计失败")
		}
	}

	return nil
}

