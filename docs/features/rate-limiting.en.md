# Rate Limiting

Website Defender includes per-IP rate limiting to protect against brute-force attacks and abuse. Two separate rate limiters operate at different levels.

## Global Rate Limiting

All incoming requests are subject to a global per-IP rate limit. When a client exceeds the configured threshold, subsequent requests receive a `429 Too Many Requests` response.

| Setting | Default | Description |
|---------|---------|-------------|
| `requests-per-minute` | `100` | Maximum requests per minute per IP |

```yaml
rate-limit:
  enabled: true
  requests-per-minute: 100
```

!!! tip "Tuning the Global Limit"
    The default of 100 requests per minute is suitable for most setups. If your protected applications serve high-frequency API calls from legitimate clients, consider increasing this value or whitelisting those IPs via the [IP Whitelist](ip-lists.md).

## Login Rate Limiting

The login endpoints (`/login` and `/admin-login`) have a **stricter** dedicated rate limit to prevent credential brute-force attacks. When a client exceeds the login rate limit, their IP is automatically locked out for a configurable duration.

| Setting | Default | Description |
|---------|---------|-------------|
| `login.requests-per-minute` | `5` | Maximum login attempts per minute per IP |
| `login.lockout-duration` | `300` | Lockout duration in seconds (5 minutes) |

```yaml
rate-limit:
  enabled: true
  login:
    requests-per-minute: 5
    lockout-duration: 300
```

!!! warning "Lockout Behavior"
    When an IP exceeds the login rate limit, it is locked out for the full `lockout-duration` period. During lockout, **all login attempts** from that IP are rejected immediately, even if the cooldown period for the rate limit window has passed.

## Configuration

The complete rate limiting configuration in `config/config.yaml`:

```yaml
rate-limit:
  enabled: true
  requests-per-minute: 100
  login:
    requests-per-minute: 5
    lockout-duration: 300
```

Set `enabled: false` to disable all rate limiting entirely.

For the full configuration reference, see [Configuration](../configuration/index.md).

## Middleware Position

Rate limiting runs as the **last middleware** before the route handler in the [middleware chain](../architecture/index.md):

```
SecurityHeaders → CORS → BodyLimit → AccessLog → GeoBlock → WAF → RateLimiter → Route Handler
```

This means requests blocked by geo-blocking or the WAF do not consume rate limit quota.
