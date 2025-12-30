#!/bin/bash

# 数据库迁移脚本

set -e

echo "开始数据库迁移..."

# 加载配置
CONFIG_FILE="${CONFIG_FILE:-configs/config.yaml}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 运行迁移
go run cmd/server/main.go migrate

echo "数据库迁移完成"

