.PHONY: build run test fmt lint clean proto docker

# 构建应用
build:
	@echo "构建应用..."
	@go build -o bin/server cmd/server/main.go

# 运行应用
run:
	@echo "运行应用..."
	@go run cmd/server/main.go

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 代码格式化
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run ./...

# 生成 protobuf 代码
proto:
	@echo "生成 protobuf 代码..."
	@protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/

# Docker 构建
docker:
	@echo "构建 Docker 镜像..."
	@docker build -t game-apps:latest .

# 安装依赖
deps:
	@echo "安装依赖..."
	@go mod download
	@go mod tidy

