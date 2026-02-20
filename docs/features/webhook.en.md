# Webhook Notifications

Website Defender can send HTTP webhook notifications when security events occur, enabling integration with external alerting and monitoring systems.

## How It Works

When a [security event](security-events.md) is triggered (e.g., an IP is auto-banned), Website Defender sends an HTTP POST request to the configured webhook URL with event details.

Webhook delivery is **asynchronous** -- it does not block the main request processing pipeline.

## Payload Format

```json
{
  "event_type": "auto_ban",
  "client_ip": "192.168.1.100",
  "reason": "excessive 4xx responses",
  "banned_for": "1h",
  "timestamp": "2026-02-20T10:30:00Z"
}
```

| Field | Description |
|-------|-------------|
| `event_type` | Type of security event (`auto_ban`, `brute_force`, `scan_detected`) |
| `client_ip` | The IP address involved |
| `reason` | Human-readable description of why the event occurred |
| `banned_for` | Duration of the ban (if applicable) |
| `timestamp` | ISO 8601 timestamp of the event |

## Event Filtering

You can configure which event types trigger webhook notifications:

```yaml
webhook:
  events:
    - auto_ban
    - brute_force
    - scan_detected
```

Only events matching the configured types will be sent. Remove event types from the list to suppress their notifications.

## Configuration

```yaml
webhook:
  # Webhook endpoint URL (empty = disabled)
  url: ""
  # Request timeout in seconds
  timeout: 5
  # Which events trigger notifications
  events:
    - auto_ban
    - brute_force
    - scan_detected
```

Webhook settings can also be configured via the admin dashboard under **System Settings**.

!!! tip "Integration Ideas"
    - Send to a **Slack** or **Discord** webhook for real-time team alerts
    - Forward to a **SIEM** system for security event correlation
    - Trigger **PagerDuty** or **Opsgenie** for on-call incident management
    - Post to a custom endpoint that updates firewall rules upstream

---

## Related Pages

- [Security Events](security-events.md) -- View all recorded security events
- [Threat Detection](threat-detection.md) -- How threats are detected
- [Configuration](../configuration/index.md) -- Full configuration reference
