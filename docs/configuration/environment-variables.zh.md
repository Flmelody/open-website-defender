# 环境变量

Website Defender 的配置分为**构建时**和**运行时**两部分。路径相关的变量在构建时通过 `.env` 文件或环境变量注入；而后端地址、Cookie 域名、端口等则在运行时通过 `config.yaml` 或操作系统环境变量配置，无需重新构建。

## 构建时环境变量

以下变量在构建时注入，嵌入到编译后的二进制文件和前端资源中。

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ROOT_PATH` | `/wall` | 根路径上下文，所有 API 路由和前端资源的前缀 |
| `ADMIN_PATH` | `/admin` | 管理后台的访问路径 |
| `GUARD_PATH` | `/guard` | 防护页/登录页的访问路径 |

### 配置方式

#### 方式一：通过 .env 文件

在项目根目录创建 `.env` 文件：

```bash
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
```

构建脚本会自动读取 `.env` 文件中的变量。

#### 方式二：通过环境变量

在执行构建脚本前设置环境变量：

```bash
export ROOT_PATH=/wall
export ADMIN_PATH=/admin
export GUARD_PATH=/guard

./scripts/build.sh
```

#### 方式三：修改构建脚本

直接修改 `scripts/build.sh` 中的默认值。

!!! warning "注意"
    这些路径变量是**构建时**变量，在编译完成后无法修改。如需更改路径，必须重新构建项目。

---

## 运行时配置

以下配置项在运行时生效，修改后只需重启服务，无需重新构建。

### 后端地址与 Cookie 域名

通过 `config/config.yaml` 中的 `wall` 部分配置：

```yaml
wall:
  # API 基础地址，用于跨域部署场景（默认：同源，通过 root-path 访问）
  backend-host: "https://example.com/wall"
  # 认证令牌的 Cookie 域名，设置后可在子域名间共享（如 ".example.com"）
  guard-domain: ".example.com"
```

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `wall.backend-host` | *（空，同源访问）* | 后端 API 地址，前端将向此地址发送 API 请求。留空时前端使用同源的 `ROOT_PATH` |
| `wall.guard-domain` | *（空）* | 认证令牌的 Cookie 域名。设置后（如 `.example.com`），`flmelody.token` Cookie 可在所有子域名间共享，实现单点登录 |

这两个配置项也可以通过操作系统环境变量 `BACKEND_HOST` 和 `GUARD_DOMAIN` 设置，环境变量的优先级高于 `config.yaml`。

### 端口

服务监听端口通过 `config/config.yaml` 中的 `server.port` 配置，或通过操作系统环境变量 `PORT` 设置：

```yaml
server:
  # 监听端口（默认：9999）
  port: 8080
```

```bash
# 或通过环境变量设置
export PORT=8080
./app
```

| 配置方式 | 默认值 | 说明 |
|----------|--------|------|
| `server.port`（config.yaml） | `9999` | 服务监听端口 |
| `PORT`（环境变量） | `9999` | 操作系统环境变量，优先级高于 config.yaml |

!!! info "运行时配置注入原理"
    后端在响应前端 HTML 页面时，会自动将 `wall` 配置注入到 `window.__APP_CONFIG__` 全局变量中。前端在初始化时读取该变量获取 `backend-host` 和 `guard-domain` 的值，因此无需在构建时指定这些参数。

    ```html
    <!-- 后端自动注入到 index.html 中 -->
    <script>
      window.__APP_CONFIG__ = {
        backendHost: "https://example.com/wall",
        guardDomain: ".example.com"
      };
    </script>
    ```

!!! tip "何时需要设置 backend-host"
    大多数部署场景下，前端与后端在同一域名下，`backend-host` 可以留空。仅在以下场景需要设置：

    - 前端和后端部署在不同域名或端口
    - 使用反向代理且前端访问的地址与后端实际地址不同
    - 跨域（CORS）部署方案

---

## 构建流程

构建脚本 `scripts/build.sh` 的工作流程：

1. 读取 `.env` 文件（如存在）
2. 将路径变量导出为 `VITE_` 前缀变量，注入前端构建
3. 依次构建 Guard 前端和 Admin 前端
4. 通过 Go 的 `-ldflags` 将路径变量注入后端编译
5. 生成最终的可执行文件 `app`

```bash
# 前端通过 VITE_ 前缀变量接收路径配置
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH

# 后端通过 ldflags 接收路径配置（不再包含 BACKEND_HOST）
go build -ldflags "\
  -X 'main.RootPath=$ROOT_PATH' \
  -X 'main.AdminPath=$ADMIN_PATH' \
  -X 'main.GuardPath=$GUARD_PATH' \
" -o app main.go
```

!!! note "与旧版本的区别"
    旧版本中 `BACKEND_HOST`、`GUARD_DOMAIN`、`PORT` 均为构建时变量，需要在构建前确定。新版本将这三者改为运行时配置，构建脚本不再传递 `BACKEND_HOST`。同一个二进制文件可以在不同环境中通过修改 `config.yaml` 灵活部署。

---

## URL 结构

使用默认配置时，应用的访问地址如下：

| 资源 | URL |
|------|-----|
| 管理后台 | `http://localhost:9999/wall/admin/` |
| 防护页 | `http://localhost:9999/wall/guard/` |
| API 根路径 | `http://localhost:9999/wall/` |
| 认证端点 | `http://localhost:9999/wall/auth` |
| 登录端点 | `http://localhost:9999/wall/login` |
| 管理员登录端点 | `http://localhost:9999/wall/admin-login` |

!!! tip "自定义路径示例"
    若设置 `ROOT_PATH=/defender`、`ADMIN_PATH=/dashboard`，管理后台地址变为 `http://localhost:9999/defender/dashboard/`。

---

## 相关页面

- [配置说明](index.md) - 运行时配置参考（config.yaml 完整说明）
- [数据库](database.md) - 多数据库配置详解
- [快速开始](../getting-started/index.md) - 构建和运行说明
- [开发指南](../development/index.md) - 从源码构建详解
