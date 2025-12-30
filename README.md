# 游戏服务后端应用

高性能游戏服务后端应用，使用 Go 语言开发，支持游戏进程管理、用户管理等核心功能。

## 功能特性

- 🎮 游戏进程管理（房间、会话、逻辑进程）
- 👤 用户管理（认证、资料、统计）
- 🚀 高性能架构（Gin + gRPC + WebSocket）
- 💾 多数据库支持（MySQL/PostgreSQL + Redis）
- 📊 监控和指标（Prometheus）
- 🔒 安全认证（JWT）

## 技术栈

- **Web 框架**: Gin
- **gRPC**: protobuf + gRPC
- **WebSocket**: Gorilla WebSocket
- **数据库**: GORM (MySQL/PostgreSQL)
- **缓存**: Redis
- **配置**: Viper
- **日志**: Zap
- **监控**: Prometheus

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+ 或 PostgreSQL 14+
- Redis 6.0+

### 配置

复制配置文件模板：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml` 配置数据库和 Redis 连接信息。

### 运行

```bash
# 安装依赖
go mod download

# 运行服务
go run cmd/server/main.go

# 或使用 Makefile
make run
```

### 构建

```bash
make build
```

## API 文档

- HTTP API: `/api/v1/*`
- gRPC: 端口 9090
- WebSocket: `/ws`
- 健康检查: `/health`
- 就绪检查: `/ready`
- 指标: `/metrics`

## 项目结构

```
game-apps/
├── cmd/server/          # 应用入口
├── internal/            # 内部代码
│   ├── api/            # API 层
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据访问
│   ├── model/          # 数据模型
│   └── middleware/     # 中间件
├── api/proto/          # protobuf 定义
├── configs/            # 配置文件
└── deployments/        # 部署配置
```

## 开发

```bash
# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint
```

## 部署

### Docker

```bash
docker build -t game-apps:latest .
docker run -p 8080:8080 -p 9090:9090 game-apps:latest
```

### Kubernetes

```bash
kubectl apply -f deployments/k8s/
```

## License

MIT

