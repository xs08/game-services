package admin

import (
	"context"
	"fmt"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/repository/mysql"
	"github.com/game-apps/internal/repository/postgres"
	"github.com/game-apps/internal/utils"
	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	userRepo UserRepository
}

// UserRepository 用户仓库接口
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, limit, offset int, keyword string, status *string) ([]*model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
}
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB, driver string) *UserService {
	var userRepo interface {
		GetByID(ctx context.Context, id uint) (*model.User, error)
		GetByUsername(ctx context.Context, username string) (*model.User, error)
		GetByEmail(ctx context.Context, email string) (*model.User, error)
		List(ctx context.Context, limit, offset int, keyword string, status *string) ([]*model.User, int64, error)
		Update(ctx context.Context, user *model.User) error
	}

	if driver == "mysql" {
		userRepo = mysql.NewUserRepository(db)
	} else {
		userRepo = postgres.NewUserRepository(db)
	}

	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserList 获取用户列表
type GetUserListRequest struct {
	Page     int
	PageSize int
	Keyword  string
	Status   *string
}

type GetUserListResponse struct {
	List     []*model.User `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func (s *UserService) GetUserList(ctx context.Context, req *GetUserListRequest) (*GetUserListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize
	users, total, err := s.userRepo.List(ctx, req.PageSize, offset, req.Keyword, req.Status)
	if err != nil {
		return nil, utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("获取用户列表失败: %v", err))
	}

	return &GetUserListResponse{
		List:     users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserDetail 获取用户详情
func (s *UserService) GetUserDetail(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewError(utils.ErrCodeNotFound, "用户不存在")
		}
		return nil, utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("获取用户详情失败: %v", err))
	}
	return user, nil
}

// UpdateUser 更新用户信息
type UpdateUserRequest struct {
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Status   *string `json:"status"`
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.NewError(utils.ErrCodeNotFound, "用户不存在")
		}
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("获取用户失败: %v", err))
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Email != nil {
		// 检查邮箱是否已被使用
		if existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email); err == nil && existingUser.ID != id {
			return utils.NewError(utils.ErrCodeInvalidInput, "邮箱已被使用")
		}
		user.Email = *req.Email
	}
	if req.Status != nil {
		statusInt := 1
		if *req.Status == "inactive" {
			statusInt = 2
		}
		user.Status = statusInt
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return utils.NewError(utils.ErrCodeInternal, fmt.Sprintf("更新用户失败: %v", err))
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, id uint, status string) error {
	req := &UpdateUserRequest{
		Status: &status,
	}
	return s.UpdateUser(ctx, id, req)
}

