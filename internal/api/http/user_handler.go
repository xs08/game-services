package http

import (
	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/service/user"
	"github.com/game-apps/internal/utils"
)

// UserHandler 用户处理器
type UserHandler struct {
	authService   *user.AuthService
	profileService *user.ProfileService
	statsService   *user.StatsService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(
	authService *user.AuthService,
	profileService *user.ProfileService,
	statsService *user.StatsService,
) *UserHandler {
	return &UserHandler{
		authService:    authService,
		profileService: profileService,
		statsService:   statsService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// RefreshToken 刷新令牌
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req user.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	resp, err := h.profileService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	var req user.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.profileService.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetStats 获取用户统计
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	resp, err := h.statsService.GetStats(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

