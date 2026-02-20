# 环境变量

Website Defender 的前端构建依赖一组**构建时**环境变量，用于配置 API 地址和路径。这些变量在构建时注入，嵌入到编译后的二进制文件中。

## 构建时环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BACKEND_HOST` | `http://localhost:9999/wall` | 后端 API 地址，前端将向此地址发送 API 请求 |
| `ROOT_PATH` | `/wall` | 根路径上下文，所有 API 路由的前缀 |
| `ADMIN_PATH` | `/admin` | 管理后台的访问路径 |
| `GUARD_PATH` | `/guard` | 防护页/登录页的访问路径 |
| `GUARD_DOMAIN` | *（空）* | 认证令牌的 Cookie 域名。设置后（如 `.example.com`），`flmelody.token` Cookie 可在所有子域名间共享，实现单点登录 |
| `PORT` | `9999` | 服务监听端口 |

## 配置方式

### 方式一：通过 .env 文件

在项目根目录创建 `.env` 文件：

```bash
BACKEND_HOST=https://example.com/wall
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
GUARD_DOMAIN=
PORT=9999
```

构建脚本会自动读取 `.env` 文件中的变量。

### 方式二：通过环境变量

在执行构建脚本前设置环境变量：

```bash
export BACKEND_HOST=https://example.com/wall
export ROOT_PATH=/wall
export ADMIN_PATH=/admin
export GUARD_PATH=/guard
export GUARD_DOMAIN=
export PORT=9999

./scripts/build.sh
```

### 方式三：修改构建脚本

直接修改 `scripts/build.sh` 中的默认值。

!!! warning "注意"
    这些变量是**构建时**变量，在编译完成后无法修改。如需更改，必须重新构建项目。

!!! info "运行时配置"
    运行时可调整的配置项（如数据库、速率限制、WAF 等）通过 `config/config.yaml` 文件管理，详见[配置说明](index.md)。

## 构建流程

构建脚本 `scripts/build.sh` 的工作流程：

1. 读取 `.env` 文件（如存在）
2. 将环境变量注入前端构建（通过 Vite 的 `VITE_` 前缀变量）
3. 依次构建 Guard 前端和 Admin 前端
4. 通过 Go 的 `-ldflags` 将变量注入后端编译
5. 生成最终的可执行文件 `app`

---

## 相关页面

- [配置说明](index.md) - 运行时配置参考
- [快速开始](../getting-started/index.md) - 构建和运行说明
- [开发指南](../development/index.md) - 从源码构建详解
