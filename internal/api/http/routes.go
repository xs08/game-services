package http

import (
	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/middleware"
	"github.com/game-apps/internal/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// SetupRoutes 设置路由
func SetupRoutes(
	router *gin.Engine,
	userHandler *UserHandler,
	gameHandler *GameHandler,
	jwtService *utils.JWTService,
	logger *zap.Logger,
) {
	// 全局中间件
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.MetricsMiddleware())

	// 健康检查
	router.GET("/health", healthCheck)
	router.GET("/ready", readyCheck)

	// Metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 用户相关（不需要认证）
		user := v1.Group("/user")
		{
			user.POST("/register", userHandler.Register)
			user.POST("/login", userHandler.Login)
			user.POST("/refresh", userHandler.RefreshToken)
		}

		// 需要认证的用户接口
		authUser := v1.Group("/user")
		authUser.Use(middleware.AuthMiddleware(jwtService))
		{
			authUser.POST("/logout", userHandler.Logout)
			authUser.GET("/profile", userHandler.GetProfile)
			authUser.PUT("/profile", userHandler.UpdateProfile)
			authUser.GET("/stats", userHandler.GetStats)
		}

		// 游戏相关（需要认证）
		game := v1.Group("/game")
		game.Use(middleware.AuthMiddleware(jwtService))
		{
			// 房间管理
			game.POST("/rooms", gameHandler.CreateRoom)
			game.POST("/rooms/join", gameHandler.JoinRoom)
			game.DELETE("/rooms/:id", gameHandler.LeaveRoom)
			game.GET("/rooms/:id", gameHandler.GetRoom)
			game.GET("/rooms", gameHandler.ListRooms)

			// 游戏进程
			game.POST("/rooms/:id/start", gameHandler.StartGame)
			game.GET("/rooms/:id/state", gameHandler.GetGameState)
		}
	}
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
	})
}

// readyCheck 就绪检查
func readyCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ready",
	})
}

