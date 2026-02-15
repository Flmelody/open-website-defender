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
# 服务器配置
# ==================================================
server:
  # 最大请求体大小（MB），默认 10
  max-body-size-mb: 10

# ==================================================
# 默认用户凭据（首次启动时创建）
# ==================================================
default-user:
  username: defender
  password: defender

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

## 相关页面

- [环境变量](environment-variables.md) - 构建时环境变量
- [数据库](database.md) - 多数据库配置详解
- [部署指南](../deployment/index.md) - 部署和运维建议
