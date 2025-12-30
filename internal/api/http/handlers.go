package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/utils"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*utils.AppError); ok {
		c.JSON(appErr.HTTPStatus(), Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
	} else {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    utils.ErrCodeInternal,
			Message: err.Error(),
		})
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

