package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/game-apps/internal/utils"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域，生产环境应该检查来源
	},
}

// HandleWebSocket WebSocket 处理器
func HandleWebSocket(hub *Hub, jwtService *utils.JWTService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从查询参数获取 Token
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    utils.ErrCodeUnauthorized,
				"message": "未提供认证令牌",
			})
			return
		}

		// 验证 Token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    utils.ErrCodeUnauthorized,
				"message": "无效的认证令牌",
			})
			return
		}

		// 升级连接
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("升级 WebSocket 连接失败", zap.Error(err))
			return
		}

		// 创建客户端
		client := &Client{
			Hub:      hub,
			Conn:     conn,
			Send:     make(chan []byte, 256),
			UserID:   claims.UserID,
			Username: claims.Username,
		}

		// 注册客户端
		hub.register <- client

		// 启动读写协程
		go client.WritePump()
		go client.ReadPump()
	}
}

