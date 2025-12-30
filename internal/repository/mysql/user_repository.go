package mysql

import (
	"context"
	"errors"

	"github.com/game-apps/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// UserProfileRepository 用户资料数据访问层
type UserProfileRepository struct {
	db *gorm.DB
}

// NewUserProfileRepository 创建用户资料仓库
func NewUserProfileRepository(db *gorm.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

// Create 创建用户资料
func (r *UserProfileRepository) Create(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// GetByUserID 根据用户 ID 获取资料
func (r *UserProfileRepository) GetByUserID(ctx context.Context, userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// Update 更新用户资料
func (r *UserProfileRepository) Update(ctx context.Context, profile *model.UserProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// UserStatsRepository 用户统计数据访问层
type UserStatsRepository struct {
	db *gorm.DB
}

// NewUserStatsRepository 创建用户统计仓库
func NewUserStatsRepository(db *gorm.DB) *UserStatsRepository {
	return &UserStatsRepository{db: db}
}

// Create 创建用户统计
func (r *UserStatsRepository) Create(ctx context.Context, stats *model.UserStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

// GetByUserID 根据用户 ID 获取统计
func (r *UserStatsRepository) GetByUserID(ctx context.Context, userID uint) (*model.UserStats, error) {
	var stats model.UserStats
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

// Update 更新用户统计
func (r *UserStatsRepository) Update(ctx context.Context, stats *model.UserStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

// UpdateWinRate 更新胜率
func (r *UserStatsRepository) UpdateWinRate(ctx context.Context, userID uint) error {
	var stats model.UserStats
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error; err != nil {
		return err
	}

	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
		return r.db.WithContext(ctx).Save(&stats).Error
	}
	return nil
}

