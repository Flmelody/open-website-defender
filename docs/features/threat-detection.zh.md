# 威胁检测

Website Defender 内置高级威胁检测引擎，可自动识别并封禁恶意行为模式。当检测到可疑活动时，攻击 IP 会被自动加入黑名单并设置可配置的封禁时长。

## 检测方式

### 4xx 错误洪泛检测

监控产生大量客户端错误响应的 IP，这通常意味着自动化扫描或探测行为。

- **默认阈值**：60 秒内产生 20 个 4xx 状态码响应
- **封禁时长**：1 小时

### 路径扫描检测

识别系统性探测不存在路径（404 响应）的 IP，这是常见的侦察技术。

- **默认阈值**：5 分钟内产生 10 个不同的 404 响应
- **封禁时长**：4 小时

### 速率限制滥用检测

捕捉反复触发速率限制的 IP，表明存在自动化滥用或拒绝服务攻击。

- **默认阈值**：5 分钟内触发 5 次速率限制
- **封禁时长**：2 小时

### 暴力破解检测

检测在所有登录端点（`/login`、`/admin-login`、Guard 登录）上有大量失败登录尝试的 IP。

- **默认阈值**：10 分钟内 10 次登录失败
- **封禁时长**：1 小时

## 威胁评分

每个 IP 会根据其行为累积动态威胁评分。评分会随时间自动衰减（1 小时 TTL）。

| 事件 | 评分 |
|------|------|
| WAF 拦截 (403) | +5 |
| 速率限制命中 (429) | +3 |
| 客户端错误 (4xx) | +1 |

威胁评分与 [JS 挑战](js-challenge.md)功能集成 -- 当 JS 挑战设置为 `suspicious` 模式时，仅对威胁评分较高的 IP 发起挑战。

!!! info "防止反馈循环"
    已被封禁的 IP 不会继续累积威胁评分，避免分数虚假膨胀。

## 自动封禁行为

当检测阈值被触发时：

1. IP 被加入黑名单并设置临时封禁（自动过期）
2. 记录一条[安全事件](security-events.md)
3. 发送 [Webhook 通知](webhook.md)（如已配置）

自动封禁的条目会标注备注信息（如 "auto-banned: excessive 4xx responses"），并包含过期时间戳。过期条目每 10 分钟自动清理一次。

## 配置

```yaml
threat-detection:
  enabled: true
  # 4xx 响应阈值
  status-code-threshold: 20
  status-code-window: 60          # 秒
  # 速率限制滥用
  rate-limit-abuse-threshold: 5
  rate-limit-abuse-window: 300    # 秒
  # 默认自动封禁时长
  auto-ban-duration: 3600         # 1 小时
  # 路径扫描检测
  scan-threshold: 10
  scan-window: 300                # 秒
  scan-ban-duration: 14400        # 4 小时
  # 暴力破解检测
  brute-force-threshold: 10
  brute-force-window: 600         # 秒
  brute-force-ban-duration: 3600  # 1 小时
```

!!! tip "调整阈值"
    建议从默认值开始，根据实际流量模式进行调整。如果在[安全事件](security-events.md)页面中发现误报，请提高阈值；对于高安全性环境，可以降低阈值。

---

## 相关页面

- [安全事件](security-events.md) -- 查看和分析检测到的威胁
- [JS 挑战](js-challenge.md) -- 对可疑 IP 的工作量证明挑战
- [Webhook 通知](webhook.md) -- 在检测到威胁时获取通知
- [IP 黑白名单](ip-lists.md) -- 手动和自动 IP 封禁
- [访问日志](access-logs.md) -- 请求日志与分析
