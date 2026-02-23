# 配置说明

Website Defender 支持通过 `config/config.yaml` 配置文件进行运行时配置。

## 运行时配置参考

以下是完整的配置文件示例及详细注释：

```yaml
# ==================================================
# 数据库配置
# 支持的驱动：sqlite（默认）、postgres、mysql
# ==================================================
database:
  driver: sqlite
  # SQLite 配置（driver 为 sqlite 时使用）
  # file-path: ./data/app.db

  # PostgreSQL 配置（driver 为 postgres 时使用）
  # host: localhost
  # port: 5432
  # name: open_website_defender
  # user: postgres
  # password: your_password
  # ssl-mode: disable

  # MySQL 配置（driver 为 mysql 时使用）
  # host: localhost
  # port: 3306
  # name: open_website_defender
  # user: root
  # password: your_password

# ==================================================
# 安全配置
# ==================================================
security:
  # JWT 签名密钥
  # 留空则每次重启时随机生成
  # 重要：生产环境中务必设置此值，否则重启后已签发的令牌将失效
  jwt-secret: ""
  # 令牌过期时间（小时），默认 24
  token-expiration-hours: 24
  # 受信任设备记忆天数（默认 7）
  trusted-device-days: 7
  # 管理员 2FA 恢复密钥（留空则禁用恢复端点）
  admin-recovery-key: ""
  # 仅允许本地访问 2FA 恢复端点（默认 true）
  admin-recovery-local-only: true

  # CORS 跨域配置
  cors:
    # 允许的源列表
    # 空列表 = 反射任意源（仅用于开发环境）
    # 生产环境中请设置明确的源：
    # allowed-origins:
    #   - "https://example.com"
    #   - "https://admin.example.com"
    allowed-origins: []
    allow-credentials: true

  # 安全响应头配置
  headers:
    # 启用 HSTS（仅在确认使用 HTTPS 时开启）
    hsts-enabled: false
    # X-Frame-Options: DENY、SAMEORIGIN 或留空禁用
    frame-options: "DENY"

# ==================================================
# 速率限制配置
# ==================================================
rate-limit:
  enabled: true
  # 全局限速：每个 IP 每分钟最大请求数
  requests-per-minute: 100
  # 登录限速（更严格）
  login:
    requests-per-minute: 5
    # 超出限制后的锁定时间（秒）
    lockout-duration: 300

# ==================================================
# WAF 请求过滤配置
# SQL 注入、XSS、路径穿越检测
# ==================================================
request-filtering:
  enabled: true
  # 语义分析引擎（更深层的 SQLi/XSS 检测）
  semantic-analysis:
    enabled: true

# ==================================================
# 地域封锁配置
# ==================================================
geo-blocking:
  enabled: false
  # MaxMind GeoLite2-Country.mmdb 文件路径
  # 下载地址：https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
  database-path: ""
  # 封锁的国家代码通过管理后台 API 管理（POST /geo-block-rules）

# ==================================================
# 缓存配置
# ==================================================
cache:
  # 最大内存缓存大小（MB），默认 100
  # size-mb: 100
  # 多实例同步轮询间隔（秒），0 = 禁用
  # sync-interval: 0

# ==================================================
# 服务器配置
# ==================================================
server:
  # 监听端口（默认 9999，也可通过 PORT 环境变量设置）
  # port: 9999
  # 最大请求体大小（MB），默认 10
  max-body-size-mb: 10

# ==================================================
# 默认用户凭据（首次启动时创建）
# ==================================================
default-user:
  username: defender
  password: defender

# ==================================================
# OAuth2/OIDC 提供者配置
# ==================================================
oauth:
  enabled: true
  issuer: "https://auth.example.com/wall"
  rsa-private-key-path: ""
  authorization-code-lifetime: 300
  access-token-lifetime: 3600
  refresh-token-lifetime: 2592000
  id-token-lifetime: 3600

# ==================================================
# 威胁检测配置（异常行为自动封禁）
# ==================================================
threat-detection:
  enabled: true
  status-code-threshold: 20
  status-code-window: 60
  rate-limit-abuse-threshold: 5
  rate-limit-abuse-window: 300
  auto-ban-duration: 3600
  scan-threshold: 10
  scan-window: 300
  scan-ban-duration: 14400
  brute-force-threshold: 10
  brute-force-window: 600
  brute-force-ban-duration: 3600

# ==================================================
# JS 挑战（工作量证明）配置
# ==================================================
js-challenge:
  enabled: false
  mode: "suspicious"    # off | suspicious | all
  difficulty: 4         # 1-6
  cookie-ttl: 86400
  cookie-secret: ""

# ==================================================
# Webhook 通知配置
# ==================================================
webhook:
  url: ""
  timeout: 5
  events:
    - auto_ban
    - brute_force
    - scan_detected

# ==================================================
# Wall（前端运行时配置）
# 这些值会在运行时注入到前端 HTML 中
# ==================================================
# wall:
#   backend-host: ""       # 跨域部署时的 API 基础 URL
#   guard-domain: ""       # SSO 跨子域名共享 Cookie 的域

# ==================================================
# 机器人管理配置
# ==================================================
bot-management:
  enabled: false
  # 对重复违规者从 JS 挑战升级到验证码
  challenge-escalation: false
  # 验证码提供商配置
  captcha:
    # 提供商：hcaptcha 或 turnstile
    provider: ""
    site-key: ""
    secret-key: ""
    cookie-ttl: 86400

# ==================================================
# 信任代理列表
# ==================================================
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

!!! warning "生产环境必要配置"
    以下配置项在生产环境中务必修改：

    - `security.jwt-secret`：设置一个强随机字符串
    - `default-user.password`：首次登录后立即修改默认密码
    - `security.cors.allowed-origins`：设置明确的允许源
    - `trustedProxies`：设置实际的信任代理 IP

!!! tip "配置热加载"
    部分配置项支持通过管理后台的"重载配置"功能或 `POST /system/reload` API 热加载，无需重启服务。

---

## 各配置段详解

### 数据库

配置数据库后端。Website Defender 支持 SQLite、PostgreSQL 和 MySQL。

详细的多数据库配置示例请参阅 [数据库](database.md)。

### 缓存

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `size-mb` | `100` | 最大内存缓存大小（MB） |
| `sync-interval` | `0`（禁用） | 多实例同步轮询间隔（秒） |

### 安全

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `jwt-secret` | `""`（随机） | JWT 令牌签名密钥 |
| `token-expiration-hours` | `24` | JWT 令牌有效期（小时） |
| `trusted-device-days` | `7` | 受信任设备记忆天数 |
| `admin-recovery-key` | `""`（禁用） | 管理员 2FA 恢复密钥 |
| `admin-recovery-local-only` | `true` | 限制 2FA 恢复仅限本地访问 |
| `cors.allowed-origins` | `[]`（宽松） | 允许的 CORS 源列表 |
| `cors.allow-credentials` | `true` | 是否允许 CORS 携带凭据 |
| `headers.hsts-enabled` | `false` | 启用 HTTP 严格传输安全 |
| `headers.frame-options` | `"DENY"` | X-Frame-Options 响应头 |

!!! warning "生产环境安全设置"
    生产环境中务必设置：

    - 稳定的 `jwt-secret` 以确保令牌在重启后仍然有效
    - 明确的 `cors.allowed-origins` 替代宽松的默认设置
    - 如使用 HTTPS，启用 `hsts-enabled: true`

### 速率限制

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `true` | 启用或禁用速率限制 |
| `requests-per-minute` | `100` | 每个 IP 全局每分钟请求限制 |
| `login.requests-per-minute` | `5` | 登录端点每个 IP 限制 |
| `login.lockout-duration` | `300` | 登录锁定时长（秒） |

更多详情请参阅 [速率限制](../features/rate-limiting.md)。

### 请求过滤（WAF）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `true` | 启用或禁用 WAF |
| `semantic-analysis.enabled` | `true` | 启用语义分析引擎进行更深层的 SQLi/XSS 检测 |

更多详情请参阅 [WAF 规则](../features/waf.md)。

### 地域封锁

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `false` | 启用或禁用地域 IP 封锁 |
| `database-path` | `""` | MaxMind GeoLite2-Country `.mmdb` 文件路径 |

更多详情请参阅 [地域 IP 封锁](../features/geo-blocking.md)。

### 服务器

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `port` | `9999` | 监听端口（也可通过 `PORT` 环境变量设置） |
| `max-body-size-mb` | `10` | 最大请求体大小（MB） |

### 默认用户

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `username` | `defender` | 首次启动时创建的默认管理员用户名 |
| `password` | `defender` | 首次启动时创建的默认管理员密码 |

!!! warning "修改默认凭据"
    首次登录后请立即修改默认用户名和密码。

### 威胁检测

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `true` | 启用或禁用自动威胁检测 |
| `status-code-threshold` | `20` | 自动封禁前的 4xx 响应次数 |
| `status-code-window` | `60` | 4xx 计数时间窗口（秒） |
| `rate-limit-abuse-threshold` | `5` | 自动封禁前的速率限制触发次数 |
| `rate-limit-abuse-window` | `300` | 速率限制计数时间窗口（秒） |
| `auto-ban-duration` | `3600` | 默认自动封禁时长（秒） |
| `scan-threshold` | `10` | 扫描检测前的 404 响应次数 |
| `scan-window` | `300` | 扫描计数时间窗口（秒） |
| `scan-ban-duration` | `14400` | 扫描检测封禁时长（秒） |
| `brute-force-threshold` | `10` | 暴力破解检测前的失败登录次数 |
| `brute-force-window` | `600` | 暴力破解计数时间窗口（秒） |
| `brute-force-ban-duration` | `3600` | 暴力破解封禁时长（秒） |

更多详情请参阅 [威胁检测](../features/threat-detection.md)。

### JS 挑战

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `false` | 启用或禁用 JS 挑战 |
| `mode` | `"suspicious"` | 挑战模式：`off`、`suspicious` 或 `all` |
| `difficulty` | `4` | 要求的前导零数量（1-6） |
| `cookie-ttl` | `86400` | 通过 Cookie 有效期（秒） |
| `cookie-secret` | `""` | Cookie 签名 HMAC 密钥 |

更多详情请参阅 [JS 挑战](../features/js-challenge.md)。

### Webhook

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `url` | `""`（禁用） | Webhook 端点 URL |
| `timeout` | `5` | 请求超时时间（秒） |
| `events` | `[auto_ban, brute_force, scan_detected]` | 触发通知的事件类型 |

更多详情请参阅 [Webhook](../features/webhook.md)。

### Wall（前端运行时配置）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `backend-host` | `""`（同源） | API 基础 URL，仅在跨域部署时需要设置 |
| `guard-domain` | `""` | SSO 跨子域名共享 Cookie 的域 |

这些值会在运行时通过 `window.__APP_CONFIG__` 注入到前端 HTML 中。大多数部署场景下无需设置，仅当前端和后端位于不同源、或需要在多个子域名之间共享 Guard Cookie 时才需要配置。

### 机器人管理

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `enabled` | `false` | 启用或禁用机器人管理 |
| `challenge-escalation` | `false` | 对重复违规者从 JS 挑战升级到验证码 |
| `captcha.provider` | `""`（禁用） | 验证码提供商：`hcaptcha` 或 `turnstile` |
| `captcha.site-key` | `""` | 验证码提供商站点密钥 |
| `captcha.secret-key` | `""` | 验证码提供商密钥 |
| `captcha.cookie-ttl` | `86400` | 验证码通过 Cookie 有效期（秒） |

### 信任代理

`trustedProxies` 列表指定了哪些代理 IP 是可信的，用于正确解析 `X-Forwarded-For` 等转发头中的客户端 IP。在反向代理后运行时需要正确配置此项。

```yaml
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

---

## 相关页面

- [环境变量](environment-variables.md) - 构建时环境变量
- [数据库](database.md) - 多数据库配置详解
- [部署指南](../deployment/index.md) - 部署和运维建议
