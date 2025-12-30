package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/service/admin"
	"github.com/game-apps/internal/service/user"
	"github.com/game-apps/internal/utils"
)

// AdminHandler 管理处理器
type AdminHandler struct {
	configService  *admin.ConfigService
	userService    *admin.UserService
	systemService  *admin.SystemService
	authService    *user.AuthService
}

// NewAdminHandler 创建管理处理器
func NewAdminHandler(
	configService *admin.ConfigService,
	userService *admin.UserService,
	systemService *admin.SystemService,
	authService *user.AuthService,
) *AdminHandler {
	return &AdminHandler{
		configService: configService,
		userService:   userService,
		systemService: systemService,
		authService:   authService,
	}
}

// AdminLogin 管理登录（复用用户登录逻辑）
func (h *AdminHandler) AdminLogin(c *gin.Context) {
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

	// 获取用户信息
	userInfo, err := h.userService.GetUserDetail(c.Request.Context(), resp.UserID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"token":         resp.Token,
		"refresh_token": resp.RefreshToken,
		"user": gin.H{
			"id":       userInfo.ID,
			"username": userInfo.Username,
			"email":    userInfo.Email,
			"nickname": userInfo.Nickname,
			"role":     "admin", // TODO: 从数据库获取实际角色
			"status":   userInfo.Status,
		},
	})
}

// GetConfig 获取服务配置
func (h *AdminHandler) GetConfig(c *gin.Context) {
	service := c.Param("service")
	if service == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "服务类型不能为空"))
		return
	}

	content, fileType, err := h.configService.GetConfig(c.Request.Context(), service)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"content":  content,
		"service":  service,
		"file_type": fileType,
	})
}

// UpdateConfig 更新服务配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	service := c.Param("service")
	if service == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "服务类型不能为空"))
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.configService.UpdateConfig(c.Request.Context(), service, req.Content); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ValidateConfig 验证配置
func (h *AdminHandler) ValidateConfig(c *gin.Context) {
	service := c.Param("service")
	if service == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "服务类型不能为空"))
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	err := h.configService.ValidateConfig(service, req.Content)
	if err != nil {
		Success(c, gin.H{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	Success(c, gin.H{
		"valid": true,
	})
}

// ReloadConfig 重新加载配置
func (h *AdminHandler) ReloadConfig(c *gin.Context) {
	service := c.Param("service")
	if service == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "服务类型不能为空"))
		return
	}

	// TODO: 实现配置热重载功能
	// 这通常需要向服务发送信号或通过管理接口触发重载
	Success(c, gin.H{
		"message": "配置重新加载请求已提交，服务可能需要重启才能生效",
	})
}

// GetUserList 获取用户列表
func (h *AdminHandler) GetUserList(c *gin.Context) {
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	keyword := c.Query("keyword")
	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	req := &admin.GetUserListRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
		Status:   statusPtr,
	}

	resp, err := h.userService.GetUserList(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// GetUserDetail 获取用户详情
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的用户ID"))
		return
	}

	user, err := h.userService.GetUserDetail(c.Request.Context(), uint(id))
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, user)
}

// UpdateUser 更新用户信息
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的用户ID"))
		return
	}

	var req admin.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的用户ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetSystemConfig 获取系统配置
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	config, err := h.systemService.GetSystemConfig(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, config)
}

// UpdateSystemConfig 更新系统配置
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	var config admin.SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.systemService.UpdateSystemConfig(c.Request.Context(), &config); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetSystemConfigCategory 获取分类配置
func (h *AdminHandler) GetSystemConfigCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "配置分类不能为空"))
		return
	}

	config, err := h.systemService.GetSystemConfigCategory(c.Request.Context(), category)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, config)
}

// UpdateSystemConfigCategory 更新分类配置
func (h *AdminHandler) UpdateSystemConfigCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "配置分类不能为空"))
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	if err := h.systemService.UpdateSystemConfigCategory(c.Request.Context(), category, data); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

