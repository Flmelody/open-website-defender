# 部署

## 概述

Website Defender 采用单文件部署模式，极大简化了部署和运维流程。

## 部署特性

### 单文件部署

前端资源（Admin 管理后台和 Guard 防护页）通过 Go 的 `go:embed` 指令嵌入到编译后的二进制文件中。部署时只需分发一个可执行文件，无需额外的静态资源目录。

```bash
# 部署只需这一个文件 + 配置
./app
```

## 快速部署

### 方式 A：下载预编译二进制

从 [GitHub Releases](https://github.com/Flmelody/open-website-defender/releases) 下载对应平台的最新版本。发布包包含二进制文件和默认 `config.yaml`：

```bash
tar -xzf open-website-defender-linux-amd64.tar.gz
cd open-website-defender-linux-amd64

# 按需编辑运行时配置
vim config/config.yaml

# 运行
./open-website-defender
```

预编译二进制使用默认路径（`/wall`、`/admin`、`/guard`）。如需自定义路径，请从源码构建（参见方式 B）。

### 方式 B：从源码构建

#### 1. 设置构建时路径（可选）

URL 路径（`ROOT_PATH`、`ADMIN_PATH`、`GUARD_PATH`）在构建时编译进二进制文件。默认值适用于大多数部署场景：

```bash
# .env（仅在需要修改默认路径时创建）
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
```

!!! note "构建时配置 vs 运行时配置"
    只有 URL 路径会在构建时写入二进制文件。`backend-host`、`guard-domain`、数据库、服务端口等设置都是运行时配置，通过 `config.yaml` 管理，修改无需重新构建。

#### 2. 构建

```bash
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender
./scripts/build.sh
```

#### 3. 多平台发布构建

构建多平台二进制（用于 GitHub Releases）：

```bash
./scripts/release.sh
```

这会在 `dist/` 目录下生成 linux/amd64、linux/arm64、darwin/amd64、darwin/arm64 和 windows/amd64 的压缩包。

### 2. 配置

创建或编辑 `config/config.yaml`：

```yaml
database:
  driver: sqlite

security:
  jwt-secret: "your-secure-random-secret"

default-user:
  username: admin
  password: "a-strong-password"

trustedProxies:
  - "127.0.0.1"

# 可选：仅在跨域部署时需要
# wall:
#   backend-host: "https://defender.example.com/wall"
```

!!! tip "配置文件位置"
    默认读取 `config/config.yaml`。确保部署目录结构为：
    ```
    /your-deploy-path/
    ├── app                    # 可执行文件
    └── config/
        └── config.yaml        # 运行时配置
    ```

### 3. 运行

```bash
./app
```

默认监听端口 `9999`。可在 `config.yaml` 中修改：

```yaml
server:
  port: 8080
```

### 4. 配置 Nginx

配置 Nginx 使用 Website Defender 作为认证提供者。完整配置指南请参阅 [Nginx 配置](nginx-setup.md)。

## 信任代理

在反向代理环境中，需要配置信任代理以正确获取客户端 IP：

```yaml
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

!!! warning "信任代理安全"
    仅将实际的反向代理 IP 加入信任列表。错误的信任代理配置可能导致 IP 伪造，影响 IP 黑白名单、速率限制和访问日志的准确性。

## 优雅关停

Website Defender 支持优雅关停（Graceful Shutdown）：

- 接收到 `SIGINT` 或 `SIGTERM` 信号时，停止接受新请求
- 等待正在处理的请求完成
- 安全关闭数据库连接和其他资源

!!! info "进程管理"
    建议使用 `systemd`、`supervisord` 或其他进程管理工具来管理 Defender 进程，确保服务的自动重启和日志收集。

## 作为系统服务运行

示例 `systemd` 单元文件：

```ini
[Unit]
Description=Website Defender WAF
After=network.target

[Service]
Type=simple
ExecStart=/opt/defender/app
WorkingDirectory=/opt/defender
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

!!! tip "工作目录"
    使用 SQLite 默认路径（`./data/app.db`）时，确保 `WorkingDirectory` 设置正确，以便数据库文件创建在预期位置。

## 部署检查清单

在将 Website Defender 部署到生产环境之前，请确认以下事项：

- [ ] 修改默认用户密码（`defender/defender`）
- [ ] 设置 `security.jwt-secret`（避免重启后令牌失效）
- [ ] 配置 `security.cors.allowed-origins`（限制跨域来源）
- [ ] 配置 `trustedProxies`（正确获取客户端 IP）
- [ ] 配置 Nginx `auth_request` 集成（参阅 [Nginx 配置](nginx-setup.md)）
- [ ] 选择合适的数据库（参阅 [配置说明](../configuration/index.md)）
- [ ] 启用 WAF 和速率限制
- [ ] 配置 HTTPS 和 HSTS（如适用）

---

## 相关页面

- [Nginx 配置](nginx-setup.md) - 详细的 Nginx 集成指南
- [配置说明](../configuration/index.md) - 完整的配置参考
