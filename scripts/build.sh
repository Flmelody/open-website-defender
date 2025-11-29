#!/bin/bash

# 构建脚本 - 从环境变量注入配置

set -e

# 读取根目录的 .env 文件
if [ -f .env ]; then
  echo "Loading .env file..."
  export $(grep -v '^#' .env | xargs)
fi

# 读取环境变量
BACKEND_HOST=${BACKEND_HOST:-"http://localhost:9999/wall"}
ROOT_PATH=${ROOT_PATH:-"/wall"}
ADMIN_PATH=${ADMIN_PATH:-"/admin"}
GUARD_PATH=${GUARD_PATH:-"/guard"}

echo "========================================="
echo "Building with configuration:"
echo "BACKEND_HOST: $BACKEND_HOST"
echo "ROOT_PATH: $ROOT_PATH"
echo "ADMIN_PATH: $ADMIN_PATH"
echo "GUARD_PATH: $GUARD_PATH"
echo "========================================="

# 1. 构建前端
echo "Building guard..."
cd ui/guard
export VITE_BACKEND_HOST=$BACKEND_HOST
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH
npm run build
cd ../..

echo "Building admin..."
cd ui/admin
export VITE_BACKEND_HOST=$BACKEND_HOST
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH
npm run build
cd ../..

# 2. 构建 Go 后端
echo "Building backend..."
go build -ldflags "\
  -X 'main.BackendHost=$BACKEND_HOST' \
  -X 'main.RootPath=$ROOT_PATH' \
  -X 'main.AdminPath=$ADMIN_PATH' \
  -X 'main.GuardPath=$GUARD_PATH' \
" -o app main.go

echo "========================================="
echo "Build completed!"
echo "Binary: ./app"
echo "Configuration is embedded in the binary"
echo "========================================="

