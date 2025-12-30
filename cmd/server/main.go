package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/api/http"
	"github.com/game-apps/internal/api/websocket"
	"github.com/game-apps/internal/config"
	"github.com/game-apps/internal/middleware"
	"github.com/game-apps/internal/repository/mysql"
	"github.com/game-apps/internal/repository/postgres"
	"github.com/game-apps/internal/repository/redis"
	"github.com/game-apps/internal/service/game"
	"github.com/game-apps/internal/service/user"
	"github.com/game-apps/internal/utils"
	"github.com/game-apps/internal/model"
	"github.com/game-apps/pkg/cache"
	"github.com/game-apps/pkg/database"
	"github.com/game-apps/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.Output,
		logger.FileConfig{
			Filename:   cfg.Log.File.Filename,
			MaxSize:    cfg.Log.File.MaxSize,
			MaxBackups: cfg.Log.File.MaxBackups,
			MaxAge:     cfg.Log.File.MaxAge,
			Compress:   cfg.Log.File.Compress,
		},
	); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Get()
	log.Info("应用启动", zap.Any("config", cfg))

	// 连接数据库
	var db *gorm.DB
	if cfg.Database.Driver == "mysql" {
		db, err = database.Connect(database.Config{
			Driver:          cfg.Database.Driver,
			MySQLConfig: database.MySQLConfig{
				Host:      cfg.Database.MySQL.Host,
				Port:      cfg.Database.MySQL.Port,
				User:      cfg.Database.MySQL.User,
				Password:  cfg.Database.MySQL.Password,
				DBName:    cfg.Database.MySQL.DBName,
				Charset:   cfg.Database.MySQL.Charset,
				ParseTime: cfg.Database.MySQL.ParseTime,
				Loc:       cfg.Database.MySQL.Loc,
			},
			MaxOpenConns:    cfg.Database.MySQL.MaxOpenConns,
			MaxIdleConns:    cfg.Database.MySQL.MaxIdleConns,
			ConnMaxLifetime: cfg.Database.MySQL.ConnMaxLifetime,
		})
	} else {
		db, err = database.Connect(database.Config{
			Driver:          cfg.Database.Driver,
			PostgresConfig: database.PostgresConfig{
				Host:     cfg.Database.Postgres.Host,
				Port:     cfg.Database.Postgres.Port,
				User:     cfg.Database.Postgres.User,
				Password: cfg.Database.Postgres.Password,
				DBName:   cfg.Database.Postgres.DBName,
				SSLMode:  cfg.Database.Postgres.SSLMode,
			},
			MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
			MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
			ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
		})
	}
	if err != nil {
		log.Fatal("连接数据库失败", zap.Error(err))
	}
	log.Info("数据库连接成功")

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 连接 Redis
	redisClient, err := cache.NewClient(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Redis.MinIdleConns,
		cfg.Redis.DialTimeout,
		cfg.Redis.ReadTimeout,
		cfg.Redis.WriteTimeout,
	)
	if err != nil {
		log.Fatal("连接 Redis 失败", zap.Error(err))
	}
	log.Info("Redis 连接成功")

	// 初始化 Repository
	var userRepo user.UserRepository
	var userProfileRepo user.UserProfileRepository
	var userStatsRepo user.UserStatsRepository
	var roomRepo game.RoomRepository
	var roomPlayerRepo game.RoomPlayerRepository

	if cfg.Database.Driver == "mysql" {
		userRepo = mysql.NewUserRepository(db)
		userProfileRepo = mysql.NewUserProfileRepository(db)
		userStatsRepo = mysql.NewUserStatsRepository(db)
		roomRepo = mysql.NewRoomRepository(db)
		roomPlayerRepo = mysql.NewRoomPlayerRepository(db)
	} else {
		userRepo = postgres.NewUserRepository(db)
		userProfileRepo = postgres.NewUserProfileRepository(db)
		userStatsRepo = postgres.NewUserStatsRepository(db)
		roomRepo = postgres.NewRoomRepository(db)
		roomPlayerRepo = postgres.NewRoomPlayerRepository(db)
	}

	redisRepo := redis.NewRepository(redisClient)
	sessionRepo := redis.NewSessionRepository(redisRepo)
	redisRoomRepo := redis.NewRoomRepository(redisRepo)
	onlineUserRepo := redis.NewOnlineUserRepository(redisRepo)
	lockRepo := redis.NewLockRepository(redisRepo)

	// 初始化服务
	jwtService := utils.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.ExpirationHours,
		cfg.JWT.RefreshExpirationHours,
	)

	authService := user.NewAuthService(
		userRepo,
		userProfileRepo,
		userStatsRepo,
		sessionRepo,
		jwtService,
		log,
	)

	profileService := user.NewProfileService(
		userRepo,
		userProfileRepo,
		log,
	)

	statsService := user.NewStatsService(
		userStatsRepo,
		log,
	)

	roomService := game.NewRoomService(
		roomRepo,
		roomPlayerRepo,
		redisRoomRepo,
		lockRepo,
		log,
		cfg.Game.Room.MaxPlayers,
		cfg.Game.Room.DefaultTimeout,
	)

	sessionService := game.NewSessionService(
		sessionRepo,
		onlineUserRepo,
		log,
		cfg.Game.Session.HeartbeatInterval,
		cfg.Game.Session.Timeout,
	)

	processService := game.NewProcessService(
		roomRepo,
		redisRoomRepo,
		lockRepo,
		log,
		"game:events",
	)

	// 初始化 HTTP 处理器
	userHandler := http.NewUserHandler(authService, profileService, statsService)
	gameHandler := http.NewGameHandler(roomService, sessionService, processService)

	// 初始化 WebSocket Hub
	wsHub := websocket.NewHub(log)
	go wsHub.Run()

	// 设置路由
	router := gin.Default()
	http.SetupRoutes(router, userHandler, gameHandler, jwtService, log)

	// WebSocket 路由
	router.GET("/ws", websocket.HandleWebSocket(wsHub, jwtService, log))

	// 创建 HTTP 服务器
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 创建 gRPC 服务器（占位，实际实现需要 protobuf 生成代码）
	grpcServer := grpc.NewServer()

	// 启动 HTTP 服务器
	go func() {
		log.Info("HTTP 服务器启动", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP 服务器启动失败", zap.Error(err))
		}
	}()

	// 启动 gRPC 服务器
	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
		log.Info("gRPC 服务器启动", zap.String("addr", grpcAddr))
		// 这里需要实现 gRPC 服务器监听
		// listener, err := net.Listen("tcp", grpcAddr)
		// if err != nil {
		// 	log.Fatal("gRPC 服务器启动失败", zap.Error(err))
		// }
		// if err := grpcServer.Serve(listener); err != nil {
		// 	log.Fatal("gRPC 服务器启动失败", zap.Error(err))
		// }
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP 服务器关闭失败", zap.Error(err))
	}

	grpcServer.GracefulStop()

	log.Info("服务器已关闭")
}

// autoMigrate 自动迁移数据库
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.UserStats{},
		&model.Room{},
		&model.RoomPlayer{},
		&model.Session{},
	)
}

