package user

import (
	"context"
	"time"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

// ProfileService 用户资料服务
type ProfileService struct {
	userRepo        UserRepository
	userProfileRepo UserProfileRepository
	logger          *zap.Logger
}

// NewProfileService 创建用户资料服务
func NewProfileService(
	userRepo UserRepository,
	userProfileRepo UserProfileRepository,
	logger *zap.Logger,
) *ProfileService {
	return &ProfileService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		logger:          logger,
	}
}

// GetProfileRequest 获取资料请求
type GetProfileRequest struct {
	UserID uint
}

// GetProfileResponse 获取资料响应
type GetProfileResponse struct {
	User    *model.User        `json:"user"`
	Profile *model.UserProfile `json:"profile"`
}

// GetProfile 获取用户资料
func (s *ProfileService) GetProfile(ctx context.Context, userID uint) (*GetProfileResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err), zap.Uint("user_id", userID))
		return nil, utils.NewError(utils.ErrCodeInternal, "获取资料失败")
	}
	if user == nil {
		return nil, utils.NewError(utils.ErrCodeNotFound, "用户不存在")
	}

	// 获取用户资料
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户资料失败", zap.Error(err), zap.Uint("user_id", userID))
		return nil, utils.NewError(utils.ErrCodeInternal, "获取资料失败")
	}

	// 如果资料不存在，创建默认资料
	if profile == nil {
		profile = &model.UserProfile{
			UserID: userID,
		}
		if err := s.userProfileRepo.Create(ctx, profile); err != nil {
			s.logger.Error("创建用户资料失败", zap.Error(err))
		}
	}

	return &GetProfileResponse{
		User:    user,
		Profile: profile,
	}, nil
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname *string    `json:"nickname"`
	Avatar   *string    `json:"avatar"`
	Gender   *int       `json:"gender"`
	Birthday *time.Time `json:"birthday"`
	Bio      *string    `json:"bio"`
	Location *string    `json:"location"`
}

// UpdateProfile 更新用户资料
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err), zap.Uint("user_id", userID))
		return utils.NewError(utils.ErrCodeInternal, "更新资料失败")
	}
	if user == nil {
		return utils.NewError(utils.ErrCodeNotFound, "用户不存在")
	}

	// 更新用户基本信息
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("更新用户失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "更新资料失败")
	}

	// 获取或创建用户资料
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户资料失败", zap.Error(err))
		return utils.NewError(utils.ErrCodeInternal, "更新资料失败")
	}

	if profile == nil {
		profile = &model.UserProfile{
			UserID: userID,
		}
	}

	// 更新资料
	if req.Gender != nil {
		profile.Gender = *req.Gender
	}
	if req.Birthday != nil {
		profile.Birthday = req.Birthday
	}
	if req.Bio != nil {
		profile.Bio = *req.Bio
	}
	if req.Location != nil {
		profile.Location = *req.Location
	}

	if profile.ID == 0 {
		if err := s.userProfileRepo.Create(ctx, profile); err != nil {
			s.logger.Error("创建用户资料失败", zap.Error(err))
			return utils.NewError(utils.ErrCodeInternal, "更新资料失败")
		}
	} else {
		if err := s.userProfileRepo.Update(ctx, profile); err != nil {
			s.logger.Error("更新用户资料失败", zap.Error(err))
			return utils.NewError(utils.ErrCodeInternal, "更新资料失败")
		}
	}

	return nil
}

