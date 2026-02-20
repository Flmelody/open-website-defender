# 部署

## 概述

Website Defender 采用单文件部署模式，极大简化了部署和运维流程。

## 部署特性

### 单文件部署

前端资源（Admin 管理后台和 Guard 防护页）通过 Go 的 `go:embed` 指令嵌入到编译后的二进制文件中。部署时只需分发一个可执行文件，无需额外的静态资源目录。

```bash
# 构建完成后，只需部署这个文件
./app
```

### 构建时环境变量（重要）

前端的 API 地址在**构建时编译进二进制文件**。部署前**必须**将 `BACKEND_HOST` 设置为实际的部署地址，否则前端无法访问后端，**登录将会失败**。

在项目根目录创建 `.env` 文件：

```bash
# 必须：修改为你的实际部署地址
BACKEND_HOST=https://defender.example.com/wall
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
GUARD_DOMAIN=
PORT=9999
```

或直接导出环境变量后再执行构建：

```bash
export BACKEND_HOST=https://defender.example.com/wall
./scripts/build.sh
```

!!! danger "常见问题"
    如果使用默认的 `BACKEND_HOST=http://localhost:9999/wall` 构建并部署到远程服务器，管理后台在用户浏览器中会尝试请求 `localhost:9999` -- API 请求将会失败，**无法登录**。构建前务必将 `BACKEND_HOST` 设置为 Defender 实例的实际访问地址。

完整的构建时变量列表请参阅[环境变量](../configuration/environment-variables.md)。

### 配置管理

运行时配置通过 `config/config.yaml` 文件管理：

- 将配置文件放置在可执行文件同级目录的 `config/` 文件夹下
- 也可通过环境变量覆盖部分配置

!!! tip "配置文件位置"
    默认读取 `config/config.yaml`。确保部署目录结构为：
    ```
    /your-deploy-path/
    ├── app                    # 可执行文件
    └── config/
        └── config.yaml        # 运行时配置
    ```

### 优雅关停

Website Defender 支持优雅关停（Graceful Shutdown）：

- 接收到 `SIGINT` 或 `SIGTERM` 信号时，停止接受新请求
- 等待正在处理的请求完成
- 安全关闭数据库连接和其他资源

!!! info "进程管理"
    建议使用 `systemd`、`supervisord` 或其他进程管理工具来管理 Defender 进程，确保服务的自动重启和日志收集。

### 信任代理

在反向代理环境中，需要配置信任代理以正确获取客户端 IP：

```yaml
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

!!! warning "信任代理安全"
    仅将实际的反向代理 IP 加入信任列表。错误的信任代理配置可能导致 IP 伪造，影响 IP 黑白名单、速率限制和访问日志的准确性。

## 部署检查清单

在将 Website Defender 部署到生产环境之前，请确认以下事项：

- [ ] 修改默认用户密码（`defender/defender`）
- [ ] 设置 `security.jwt-secret`（避免重启后令牌失效）
- [ ] 配置 `security.cors.allowed-origins`（限制跨域来源）
- [ ] 配置 `trustedProxies`（正确获取客户端 IP）
- [ ] 配置 Nginx `auth_request` 集成（参阅 [Nginx 配置](nginx-setup.md)）
- [ ] 选择合适的数据库（参阅 [数据库配置](../configuration/database.md)）
- [ ] 启用 WAF 和速率限制
- [ ] 配置 HTTPS 和 HSTS（如适用）

---

## 相关页面

- [Nginx 配置](nginx-setup.md) - 详细的 Nginx 集成指南
- [配置说明](../configuration/index.md) - 完整的配置参考
- [快速开始](../getting-started/index.md) - 构建和运行说明
