package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/utils"
)

// AdminMiddleware 管理员权限中间件
// 注意：这个中间件需要在 AuthMiddleware 之后使用
// 目前简化实现，实际应该从数据库查询用户角色
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID（由 AuthMiddleware 设置）
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    utils.ErrCodeForbidden,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		// TODO: 从数据库查询用户角色，检查是否为管理员
		// 目前简化实现，允许所有已认证用户访问管理接口
		// 在生产环境中应该实现真正的角色检查
		_ = userID

		c.Next()
	}
}

