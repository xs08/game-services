package user

import (
	"context"
	"errors"
	"time"

	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/repository/mysql"
	"github.com/game-apps/internal/repository/postgres"
	"github.com/game-apps/internal/repository/redis"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo        UserRepository
	userProfileRepo UserProfileRepository
	userStatsRepo   UserStatsRepository
	sessionRepo     *redis.SessionRepository
	jwtService      *utils.JWTService
	logger          *zap.Logger
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

// UserProfileRepository 用户资料仓库接口
type UserProfileRepository interface {
	Create(ctx context.Context, profile *model.UserProfile) error
	GetByUserID(ctx context.Context, userID uint) (*model.UserProfile, error)
	Update(ctx context.Context, profile *model.UserProfile) error
}

// UserStatsRepository 用户统计仓库接口
type UserStatsRepository interface {
	Create(ctx context.Context, stats *model.UserStats) error
	GetByUserID(ctx context.Context, userID uint) (*model.UserStats, error)
	Update(ctx context.Context, stats *model.UserStats) error
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo UserRepository,
	userProfileRepo UserProfileRepository,
	userStatsRepo UserStatsRepository,
	sessionRepo *redis.SessionRepository,
	jwtService *utils.JWTService,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userStatsRepo:   userStatsRepo,
		sessionRepo:     sessionRepo,
		jwtService:      jwtService,
		logger:          logger,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 验证用户名
	if !utils.ValidateUsername(req.Username) {
		return nil, utils.NewError(utils.ErrCodeInvalidInput, "用户名格式无效")
	}

	// 验证邮箱
	if !utils.ValidateEmail(req.Email) {
		return nil, utils.NewError(utils.ErrCodeInvalidInput, "邮箱格式无效")
	}

	// 验证密码
	if !utils.ValidatePassword(req.Password) {
		return nil, utils.NewError(utils.ErrCodeInvalidInput, "密码强度不足，需要包含大小写字母、数字和特殊字符")
	}

	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "注册失败")
	}
	if existingUser != nil {
		return nil, utils.NewError(utils.ErrCodeConflict, "用户名已存在")
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "注册失败")
	}
	if existingEmail != nil {
		return nil, utils.NewError(utils.ErrCodeConflict, "邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "注册失败")
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("创建用户失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "注册失败")
	}

	// 创建用户资料
	profile := &model.UserProfile{
		UserID: user.ID,
	}
	if err := s.userProfileRepo.Create(ctx, profile); err != nil {
		s.logger.Error("创建用户资料失败", zap.Error(err))
	}

	// 创建用户统计
	stats := &model.UserStats{
		UserID: user.ID,
	}
	if err := s.userStatsRepo.Create(ctx, stats); err != nil {
		s.logger.Error("创建用户统计失败", zap.Error(err))
	}

	// 生成 Token
	token, err := s.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "注册失败")
	}

	return &RegisterResponse{
		UserID: user.ID,
		Token:  token,
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "登录失败")
	}
	if user == nil {
		return nil, utils.NewError(utils.ErrCodeUnauthorized, "用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, utils.NewError(utils.ErrCodeForbidden, "用户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, utils.NewError(utils.ErrCodeUnauthorized, "用户名或密码错误")
	}

	// 生成 Token
	token, err := s.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "登录失败")
	}

	// 生成刷新 Token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		s.logger.Error("生成刷新 Token 失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "登录失败")
	}

	// 保存会话到 Redis
	sessionData := map[string]interface{}{
		"user_id":       user.ID,
		"username":      user.Username,
		"last_activity": time.Now().Unix(),
	}
	if err := s.sessionRepo.SetSession(ctx, user.ID, sessionData, 24*time.Hour); err != nil {
		s.logger.Warn("保存会话失败", zap.Error(err))
	}

	return &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新 Token 响应
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken 刷新 Token
func (s *AuthService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// 验证刷新 Token
	claims, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, utils.NewError(utils.ErrCodeUnauthorized, "无效的刷新令牌")
	}

	// 生成新的 Token
	token, err := s.jwtService.GenerateToken(claims.UserID, claims.Username)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "刷新令牌失败")
	}

	// 生成新的刷新 Token
	refreshToken, err := s.jwtService.GenerateRefreshToken(claims.UserID, claims.Username)
	if err != nil {
		s.logger.Error("生成刷新 Token 失败", zap.Error(err))
		return nil, utils.NewError(utils.ErrCodeInternal, "刷新令牌失败")
	}

	return &RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, userID uint) error {
	return s.sessionRepo.DeleteSession(ctx, userID)
}

// ValidateToken 验证 Token
func (s *AuthService) ValidateToken(token string) (*utils.JWTClaims, error) {
	return s.jwtService.ValidateToken(token)
}

