# Configuration

Website Defender is configured at runtime via the `config/config.yaml` file. This page documents all available configuration options.

## Full Configuration Reference

Below is the complete configuration file with all available options and their default values:

```yaml
# Database configuration
# Supported drivers: sqlite (default), postgres, mysql
database:
  driver: sqlite
  # SQLite settings (used when driver is sqlite)
  # file-path: ./data/app.db

  # PostgreSQL settings (used when driver is postgres)
  # host: localhost
  # port: 5432
  # name: open_website_defender
  # user: postgres
  # password: your_password
  # ssl-mode: disable

  # MySQL settings (used when driver is mysql)
  # host: localhost
  # port: 3306
  # name: open_website_defender
  # user: root
  # password: your_password

# Security configuration
security:
  # JWT secret key for token signing.
  # If empty, a random key is generated on each restart.
  # IMPORTANT: Set this in production to persist tokens across restarts.
  jwt-secret: ""
  # Token expiration time in hours (default: 24)
  token-expiration-hours: 24

  # CORS configuration
  cors:
    # Allowed origins. Empty list = reflect any origin (permissive, for dev only).
    # In production, set explicit origins:
    # allowed-origins:
    #   - "https://example.com"
    #   - "https://admin.example.com"
    allowed-origins: []
    allow-credentials: true

  # Security response headers
  headers:
    # Enable HSTS (only enable if behind HTTPS)
    hsts-enabled: false
    # X-Frame-Options: DENY, SAMEORIGIN, or empty to disable
    frame-options: "DENY"

# Rate limiting configuration
rate-limit:
  enabled: true
  # Global rate limit: max requests per minute per IP
  requests-per-minute: 100
  # Login-specific rate limit (stricter)
  login:
    requests-per-minute: 5
    # Lockout duration in seconds after exceeding login rate limit
    lockout-duration: 300

# Request filtering configuration (WAF: SQLi, XSS, Path Traversal detection)
request-filtering:
  enabled: true

# Geo-IP blocking configuration
geo-blocking:
  enabled: false
  # Path to MaxMind GeoLite2-Country.mmdb file
  # Download from: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
  database-path: ""
  # Blocked countries are managed via the admin API (POST /geo-block-rules)

# Server configuration
server:
  # Maximum request body size in MB (default: 10)
  max-body-size-mb: 10

# Default user credentials (created on first startup)
default-user:
  username: defender
  password: defender

# OAuth2/OIDC Provider configuration
oauth:
  enabled: true
  issuer: "https://auth.example.com/wall"
  rsa-private-key-path: ""
  authorization-code-lifetime: 300
  access-token-lifetime: 3600
  refresh-token-lifetime: 2592000
  id-token-lifetime: 3600

# Threat detection (auto-ban on anomalous behavior)
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

# JS Challenge (Proof-of-Work)
js-challenge:
  enabled: false
  mode: "suspicious"    # off | suspicious | all
  difficulty: 4         # 1-6
  cookie-ttl: 86400
  cookie-secret: ""

# Webhook notifications
webhook:
  url: ""
  timeout: 5
  events:
    - auto_ban
    - brute_force
    - scan_detected

# Trusted proxy IPs (for correct client IP detection behind reverse proxies)
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

## Configuration Sections

### Database

Configures the database backend. Website Defender supports SQLite, PostgreSQL, and MySQL.

For detailed database configuration with examples for each driver, see [Database](database.md).

### Security

| Setting | Default | Description |
|---------|---------|-------------|
| `jwt-secret` | `""` (random) | Secret key for JWT token signing |
| `token-expiration-hours` | `24` | JWT token validity period in hours |
| `admin-recovery-key` | `""` (disabled) | Recovery key for resetting admin 2FA |
| `admin-recovery-local-only` | `true` | Restrict 2FA recovery to localhost |
| `cors.allowed-origins` | `[]` (permissive) | List of allowed CORS origins |
| `cors.allow-credentials` | `true` | Allow credentials in CORS requests |
| `headers.hsts-enabled` | `false` | Enable HTTP Strict Transport Security |
| `headers.frame-options` | `"DENY"` | X-Frame-Options header value |

!!! warning "Production Security Settings"
    In production, always set:

    - A stable `jwt-secret` to persist tokens across restarts
    - Explicit `cors.allowed-origins` instead of the permissive default
    - `hsts-enabled: true` if serving over HTTPS

### Rate Limiting

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable or disable rate limiting |
| `requests-per-minute` | `100` | Global per-IP request limit |
| `login.requests-per-minute` | `5` | Login endpoint per-IP limit |
| `login.lockout-duration` | `300` | Login lockout duration in seconds |

For more details, see [Rate Limiting](../features/rate-limiting.md).

### Request Filtering (WAF)

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable or disable the WAF |

For more details, see [WAF Rules](../features/waf.md).

### Geo-Blocking

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `false` | Enable or disable geo-IP blocking |
| `database-path` | `""` | Path to MaxMind GeoLite2-Country `.mmdb` file |

For more details, see [Geo-IP Blocking](../features/geo-blocking.md).

### Server

| Setting | Default | Description |
|---------|---------|-------------|
| `max-body-size-mb` | `10` | Maximum request body size in megabytes |

### Default User

| Setting | Default | Description |
|---------|---------|-------------|
| `username` | `defender` | Default admin username created on first startup |
| `password` | `defender` | Default admin password created on first startup |

!!! warning "Change Default Credentials"
    Change the default username and password immediately after first login.

### Threat Detection

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `true` | Enable or disable automatic threat detection |
| `status-code-threshold` | `20` | 4xx responses before auto-ban |
| `status-code-window` | `60` | Time window for 4xx counting (seconds) |
| `rate-limit-abuse-threshold` | `5` | Rate limit hits before auto-ban |
| `rate-limit-abuse-window` | `300` | Time window for rate limit counting (seconds) |
| `auto-ban-duration` | `3600` | Default auto-ban duration (seconds) |
| `scan-threshold` | `10` | 404 responses before scan detection |
| `scan-window` | `300` | Time window for scan counting (seconds) |
| `scan-ban-duration` | `14400` | Scan detection ban duration (seconds) |
| `brute-force-threshold` | `10` | Failed logins before brute force detection |
| `brute-force-window` | `600` | Time window for brute force counting (seconds) |
| `brute-force-ban-duration` | `3600` | Brute force ban duration (seconds) |

For more details, see [Threat Detection](../features/threat-detection.md).

### JS Challenge

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `false` | Enable or disable JS Challenge |
| `mode` | `"suspicious"` | Challenge mode: `off`, `suspicious`, or `all` |
| `difficulty` | `4` | Number of leading zeros required (1-6) |
| `cookie-ttl` | `86400` | Pass cookie lifetime in seconds |
| `cookie-secret` | `""` | HMAC secret for cookie signing |

For more details, see [JS Challenge](../features/js-challenge.md).

### Webhook

| Setting | Default | Description |
|---------|---------|-------------|
| `url` | `""` (disabled) | Webhook endpoint URL |
| `timeout` | `5` | Request timeout in seconds |
| `events` | `[auto_ban, brute_force, scan_detected]` | Event types that trigger notifications |

For more details, see [Webhook](../features/webhook.md).

### Trusted Proxies

The `trustedProxies` list specifies which proxy IPs are trusted for forwarding headers like `X-Forwarded-For`. This ensures correct client IP detection when running behind a reverse proxy.

```yaml
trustedProxies:
  - "127.0.0.1"
  - "::1"
```

## Environment Variables

Build-time environment variables are documented separately. See [Environment Variables](environment-variables.md).
