# Webhook 通知

Website Defender 可以在安全事件发生时发送 HTTP Webhook 通知，实现与外部告警和监控系统的集成。

## 工作原理

当[安全事件](security-events.md)被触发时（例如 IP 被自动封禁），Website Defender 会向配置的 Webhook URL 发送包含事件详情的 HTTP POST 请求。

Webhook 投递是**异步**的 -- 不会阻塞主请求处理流程。

## 请求载荷格式

```json
{
  "event_type": "auto_ban",
  "client_ip": "192.168.1.100",
  "reason": "excessive 4xx responses",
  "banned_for": "1h",
  "timestamp": "2026-02-20T10:30:00Z"
}
```

| 字段 | 说明 |
|------|------|
| `event_type` | 安全事件类型（`auto_ban`、`brute_force`、`scan_detected`） |
| `client_ip` | 相关 IP 地址 |
| `reason` | 事件原因的可读描述 |
| `banned_for` | 封禁时长（如适用） |
| `timestamp` | 事件的 ISO 8601 时间戳 |

## 事件过滤

可以配置哪些事件类型触发 Webhook 通知：

```yaml
webhook:
  events:
    - auto_ban
    - brute_force
    - scan_detected
```

仅匹配配置类型的事件会被发送。从列表中移除事件类型即可屏蔽其通知。

## 配置

```yaml
webhook:
  # Webhook 端点 URL（留空则禁用）
  url: ""
  # 请求超时时间（秒）
  timeout: 5
  # 触发通知的事件类型
  events:
    - auto_ban
    - brute_force
    - scan_detected
```

Webhook 设置也可以通过管理后台的**系统设置**页面进行配置。

!!! tip "集成方案"
    - 发送到 **Slack** 或 **Discord** Webhook 实现团队实时告警
    - 转发到 **SIEM** 系统进行安全事件关联分析
    - 触发 **PagerDuty** 或 **Opsgenie** 用于值班事件管理
    - 推送到自定义端点更新上游防火墙规则

---

## 相关页面

- [安全事件](security-events.md) -- 查看所有记录的安全事件
- [威胁检测](threat-detection.md) -- 威胁如何被检测
- [配置说明](../configuration/index.md) -- 完整配置参考
