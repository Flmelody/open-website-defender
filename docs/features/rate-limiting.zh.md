# 速率限制

## 概述

Website Defender 提供两级速率限制机制，有效防止暴力破解和滥用行为。

## 全局速率限制

对所有 API 端点生效，限制每个 IP 的每分钟请求数。

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `requests-per-minute` | 100 | 每个 IP 每分钟最大请求数 |

超出限制的请求将收到 `429 Too Many Requests` 响应。

## 登录速率限制

针对登录端点（`/login` 和 `/admin-login`）的更严格限制，防止暴力破解攻击。

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `login.requests-per-minute` | 5 | 每个 IP 每分钟最大登录尝试次数 |
| `login.lockout-duration` | 300 | 超出限制后的锁定时间（秒），默认 5 分钟 |

!!! warning "登录锁定机制"
    当某个 IP 的登录尝试次数超过限制后，该 IP 将被自动锁定指定时间。锁定期间，来自该 IP 的所有登录请求将被拒绝，即使提供了正确的凭据。

## 配置

在 `config/config.yaml` 中配置速率限制：

```yaml
rate-limit:
  enabled: true
  # 全局限速：每个 IP 每分钟最大请求数
  requests-per-minute: 100
  # 登录限速（更严格）
  login:
    requests-per-minute: 5
    # 超出限制后的锁定时间（秒）
    lockout-duration: 300
```

!!! tip "调整建议"
    - 全局限速应根据实际业务需求调整，API 密集型应用可适当提高
    - 登录限速建议保持较低值（3-10次/分钟），有效防止暴力破解
    - 锁定时间可根据安全要求适当延长

!!! note "启用与禁用"
    将 `enabled` 设置为 `false` 可完全关闭速率限制功能。**不建议在生产环境中禁用此功能。**

---

## 相关页面

- [配置说明](../configuration/index.md) - 完整的运行时配置参考
- [架构说明](../architecture/index.md) - 速率限制在中间件链中的位置
- [访问日志](access-logs.md) - 查看被限速的请求记录
