# Docker 部署

Website Defender 提供了多阶段构建的 Dockerfile，可构建包含所有前端资源和 Go 二进制文件的最小化容器镜像。

## 快速开始

### 使用 Docker Compose（推荐）

```bash
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender

# 复制并编辑环境变量
cp .env.example .env
vim .env

# 按需编辑运行时配置
vim config/config.yaml

# 构建并启动
docker compose up -d
```

服务将在 `http://localhost:9999` 上可用。

Docker Compose 会自动读取项目根目录的 `.env` 文件，并将变量作为构建参数传递给 Dockerfile，确保所有配置集中管理。

### 直接使用 Docker

```bash
# 构建镜像（使用 Dockerfile ARG 默认值）
docker build -t defender .

# 或使用自定义构建参数
docker build \
  --build-arg ROOT_PATH=/api \
  -t defender .

# 运行容器（无需挂载任何卷即可使用默认配置）
docker run -d \
  --name defender \
  -p 9999:9999 \
  defender
```

## 构建参数

Dockerfile 接受以下构建参数来自定义前端配置。这些值会通过 Vite 嵌入前端资源，并通过 ldflags 嵌入 Go 二进制文件。

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `ROOT_PATH` | `/wall` | API 路由前缀 |
| `ADMIN_PATH` | `/admin` | 管理后台路径 |
| `GUARD_PATH` | `/guard` | 防护页路径 |

!!! tip "BACKEND_HOST 已改为运行时配置"
    `BACKEND_HOST` 不再是构建参数，而是通过 `config.yaml` 中的 `wall.backend-host` 或 `BACKEND_HOST` 环境变量在运行时配置。修改此项**无需**重新构建镜像。

### 配置流向

构建参数（`ROOT_PATH`、`ADMIN_PATH`、`GUARD_PATH`）与 `scripts/build.sh` 使用的 `.env` 文件共享相同的变量名：

```
.env  ──▶  docker-compose.yml (${VAR:-default})  ──▶  Dockerfile ARG
                                                         ▼
scripts/build.sh  ◀── .env                        Vite 环境变量 + Go ldflags

config.yaml (wall.backend-host)  ──▶  运行时配置（无需重新构建）
```

- **`docker compose build`**：自动读取 `.env`，将变量作为构建参数传入，覆盖 Dockerfile ARG 默认值。
- **`docker build`**（独立使用）：直接使用 Dockerfile ARG 默认值。通过 `--build-arg` 覆盖。
- **`BACKEND_HOST`**：在运行时通过 `config.yaml`（`wall.backend-host`）或 `BACKEND_HOST` 环境变量配置，不参与构建流程。

## 端口配置

应用端口可在运行时通过 `PORT` 环境变量或 `config.yaml` 中的 `server.port` 配置。使用 Docker Compose 时，`.env` 中的 `PORT` 同时驱动端口映射和应用监听端口：

```
.env (PORT=8080)
  ├──▶ 端口映射   ──▶ "8080:8080"
  └──▶ 环境变量   ──▶ Go 应用监听 :8080
```

如需更改端口，在 `.env` 中设置 `PORT`：

```bash
# .env
PORT=8080
```

然后重启：

```bash
docker compose up -d
```

也可以在 `config.yaml` 中设置端口：

```yaml
server:
  port: 8080
```

更改端口无需重新构建镜像。如果使用非标准端口且前端从不同来源访问，请在 `config.yaml` 中设置 `wall.backend-host` 以匹配。

## 数据卷

容器**无需挂载任何卷即可直接运行**——默认配置和 SQLite 数据库已内置于镜像中。卷挂载是可选的，仅在需要持久化数据或覆盖默认配置时使用。

| 路径 | 用途 | 是否必需 |
|------|------|----------|
| `/app/data` | SQLite 数据库（`app.db`）及其他持久化数据 | 否（生产环境建议挂载） |
| `/app/config` | 运行时配置（`config.yaml`） | 否（镜像内置默认配置） |

```bash
# 最简运行：不挂载卷，使用内置默认配置
docker run -d -p 9999:9999 defender

# 生产环境：挂载 data 持久化数据，挂载 config 自定义配置
docker run -d \
  -p 9999:9999 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config:/app/config \
  defender
```

!!! warning "数据持久化"
    不挂载 `/app/data` 时，SQLite 数据库存在于容器内部，容器删除后数据将**丢失**。生产环境请务必挂载此路径。

## 配置

镜像内置了默认的 `config/config.yaml`。如需自定义，挂载你自己的配置文件：

```bash
docker run -d \
  -v $(pwd)/config:/app/config \
  -p 9999:9999 \
  defender
```

配置文件格式与裸机部署完全相同，详见 [配置说明](../configuration/index.md)。

## 使用 PostgreSQL

生产环境建议使用 PostgreSQL 替代 SQLite。

1. 取消 `docker-compose.yml` 中 `postgres` 服务的注释
2. 更新 `config/config.yaml`：

```yaml
database:
  driver: postgres
  host: postgres
  port: 5432
  name: open_website_defender
  user: postgres
  password: changeme
```

3. 启动所有服务：

```bash
docker compose up -d
```

## Docker 网络与客户端 IP 检测

!!! danger "IP 相关功能需要反向代理或 Host 网络模式"
    使用 Docker 默认桥接网络加端口映射（`-p 9999:9999`）时，**所有请求的客户端 IP 都会显示为 Docker 网关 IP**（如 `172.19.0.1`），而非真实客户端 IP。这是因为 Docker 端口映射是四层 NAT 转发，不会设置任何 HTTP 头。

    这会影响**所有基于 IP 的功能**：

    | 功能 | 影响 |
    |------|------|
    | IP 黑名单 | 封禁外部 IP 无效 |
    | IP 白名单 | 放行外部 IP 无效 |
    | 速率限制 | 所有用户共享一个 IP 配额，容易误触发 |
    | 访问日志 | 所有日志显示网关 IP，非真实客户端 |
    | 地域封锁 | 无法判断真实客户端位置 |

    封禁网关 IP（如 `172.19.0.1`）会**阻断所有请求**。

### 方案一：Nginx 反向代理（推荐）

在容器前部署 Nginx，由 Nginx 设置 `X-Forwarded-For` 头携带真实客户端 IP：

```
客户端 (1.2.3.4) ──▶ Nginx ──▶ Docker 容器
                       │
                       └── X-Forwarded-For: 1.2.3.4
```

然后在 `config.yaml` 中将 Nginx/Docker 网关 IP 加入 `trustedProxies`：

```yaml
trustedProxies:
  - "172.16.0.0/12"   # Docker 网段
  - "127.0.0.1"
```

完整配置请参阅 [Nginx 配置](nginx-setup.md)。

### 方案二：Host 网络模式

使用 `network_mode: host`，容器直接使用宿主机网络栈，可直接获取真实客户端 IP：

```yaml
# docker-compose.yml
services:
  defender:
    # ...
    network_mode: host
```

此模式下无需配置 `trustedProxies`。注意使用 host 网络时 `ports` 映射会被忽略——应用直接绑定宿主机端口。

## 生产环境建议

!!! tip "生产环境检查清单"
    - 在 `config.yaml` 中设置稳定的 `security.jwt-secret`
    - 修改默认凭据（`defender/defender`）
    - 使用 PostgreSQL 或 MySQL 以获得更好的并发性能
    - 配置 `trustedProxies` 包含反向代理的 IP
    - 设置明确的 `security.cors.allowed-origins`
    - 使用命名 Docker 卷或绑定挂载 `/app/data`
    - 跨域部署时在 `config.yaml` 中设置 `wall.backend-host`

### 在 Nginx 后运行

在 Nginx 后运行 Docker 容器时，需将 Docker 网络网关或 Nginx 主机 IP 添加到 `trustedProxies`：

```yaml
trustedProxies:
  - "172.17.0.1"   # Docker 默认桥接网关
  - "127.0.0.1"
```

完整的反向代理配置请参阅 [Nginx 配置](nginx-setup.md)。

### 健康检查

在 Docker Compose 配置中添加健康检查：

```yaml
services:
  defender:
    # ...
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9999/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

### 资源限制

```yaml
services:
  defender:
    # ...
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "1.0"
```
