package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// 检查是否是 AppError
				if appErr, ok := err.(*utils.AppError); ok {
					c.JSON(appErr.HTTPStatus(), gin.H{
						"code":    appErr.Code,
						"message": appErr.Message,
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    utils.ErrCodeInternal,
						"message": "内部服务器错误",
					})
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}

