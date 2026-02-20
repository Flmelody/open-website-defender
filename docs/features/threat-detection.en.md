# Threat Detection

Website Defender includes an advanced threat detection engine that automatically identifies and blocks malicious behavior patterns. When suspicious activity is detected, the offending IP is automatically added to the blacklist with a configurable ban duration.

## Detection Methods

### 4xx Error Flood Detection

Monitors for IPs generating excessive client error responses, which often indicates automated scanning or probing.

- **Default threshold**: 20 responses with 4xx status codes within 60 seconds
- **Ban duration**: 1 hour

### Path Scan Detection

Identifies IPs that systematically probe for non-existent paths (404 responses), a common reconnaissance technique.

- **Default threshold**: 10 distinct 404 responses within 5 minutes
- **Ban duration**: 4 hours

### Rate Limit Abuse Detection

Catches IPs that repeatedly hit rate limits, indicating automated abuse or denial-of-service attempts.

- **Default threshold**: 5 rate limit hits within 5 minutes
- **Ban duration**: 2 hours

### Brute Force Detection

Detects IPs with excessive failed login attempts across all login endpoints (`/login`, `/admin-login`, guard login).

- **Default threshold**: 10 failed logins within 10 minutes
- **Ban duration**: 1 hour

## Threat Scoring

Each IP accumulates a dynamic threat score based on its behavior. Scores decay automatically over time (1-hour TTL).

| Event | Score |
|-------|-------|
| WAF block (403) | +5 |
| Rate limit hit (429) | +3 |
| Client error (4xx) | +1 |

The threat score integrates with the [JS Challenge](js-challenge.md) feature -- when JS Challenge is set to `suspicious` mode, only IPs with elevated threat scores are challenged.

!!! info "Feedback Loop Prevention"
    Already-banned IPs are excluded from threat score increments to prevent artificial score inflation.

## Auto-Ban Behavior

When a detection threshold is triggered:

1. The IP is added to the blacklist with a temporary ban (auto-expiring)
2. A [security event](security-events.md) is recorded
3. A [webhook notification](webhook.md) is sent (if configured)

Auto-banned entries are labeled with a remark (e.g., "auto-banned: excessive 4xx responses") and include an expiration timestamp. Expired entries are automatically cleaned up every 10 minutes.

## Configuration

```yaml
threat-detection:
  enabled: true
  # 4xx response threshold
  status-code-threshold: 20
  status-code-window: 60          # seconds
  # Rate limit abuse
  rate-limit-abuse-threshold: 5
  rate-limit-abuse-window: 300    # seconds
  # Default auto-ban duration
  auto-ban-duration: 3600         # 1 hour
  # Path scan detection
  scan-threshold: 10
  scan-window: 300                # seconds
  scan-ban-duration: 14400        # 4 hours
  # Brute force detection
  brute-force-threshold: 10
  brute-force-window: 600         # seconds
  brute-force-ban-duration: 3600  # 1 hour
```

!!! tip "Tuning Thresholds"
    Start with the default values and adjust based on your traffic patterns. If you see false positives in the [Security Events](security-events.md) page, increase the thresholds. For high-security environments, lower them.

---

## Related Pages

- [Security Events](security-events.md) -- View and analyze detected threats
- [JS Challenge](js-challenge.md) -- Proof-of-Work challenge for suspicious IPs
- [Webhook](webhook.md) -- Get notified when threats are detected
- [IP Lists](ip-lists.md) -- Manual and automatic IP blocking
- [Access Logs](access-logs.md) -- Request logging and analytics
