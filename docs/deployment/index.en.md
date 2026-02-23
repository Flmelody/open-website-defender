# Deployment

Website Defender is designed for simple, single-binary deployment with minimal dependencies.

## Deployment Model

| Feature | Description |
|---------|-------------|
| **Single binary** | Frontend assets (admin dashboard and guard page) are embedded via Go's `go:embed` |
| **Configuration** | Via `config/config.yaml` for runtime settings; `.env` for build-time paths |
| **Graceful shutdown** | Handles `SIGINT`/`SIGTERM` signals for clean shutdown |
| **Trusted proxies** | Configurable list of proxy IPs for correct client IP detection |
| **Database** | SQLite (default, zero-config) or PostgreSQL/MySQL for production |

## Quick Deployment

### Option A: Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/Flmelody/open-website-defender/releases). The release archive contains the binary and a default `config.yaml`:

```bash
tar -xzf open-website-defender-linux-amd64.tar.gz
cd open-website-defender-linux-amd64

# Edit runtime config as needed
vim config/config.yaml

# Run
./open-website-defender
```

Pre-built binaries use the default paths (`/wall`, `/admin`, `/guard`). If you need custom paths, build from source (see Option B).

### Option B: Build from Source

#### 1. Set Build-Time Paths (Optional)

The URL paths (`ROOT_PATH`, `ADMIN_PATH`, `GUARD_PATH`) are compiled into the binary at build time. The defaults work for most deployments:

```bash
# .env (only needed if changing default paths)
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
```

!!! note "Build-Time vs Runtime"
    Only URL paths are baked into the binary. Settings like `backend-host`, `guard-domain`, database, and server port are all runtime configuration in `config.yaml` — no rebuild needed to change them.

#### 2. Build the Binary

```bash
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender
./scripts/build.sh
```

#### 3. Multi-Platform Release Build

To build for multiple platforms (for GitHub releases):

```bash
./scripts/release.sh
```

This produces archives for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64 in the `dist/` directory.

### 2. Configure

Create or edit `config/config.yaml`:

```yaml
database:
  driver: sqlite

security:
  jwt-secret: "your-secure-random-secret"

default-user:
  username: admin
  password: "a-strong-password"

trustedProxies:
  - "127.0.0.1"

# Optional: only needed for cross-origin deployments
# wall:
#   backend-host: "https://defender.example.com/wall"
```

!!! warning "Production Checklist"
    Before deploying to production, ensure you have:

    - Set a stable `jwt-secret`
    - Changed the default user credentials
    - Configured `trustedProxies` to include your Nginx server IP(s)
    - Set explicit CORS `allowed-origins`
    - Enabled HSTS if serving over HTTPS

For the full configuration reference, see [Configuration](../configuration/index.md).

### 3. Run

```bash
./app
```

The application listens on port `9999` by default. Change it in `config.yaml`:

```yaml
server:
  port: 8080
```

### 4. Configure Nginx

Set up Nginx to use Website Defender as the auth provider. See [Nginx Setup](nginx-setup.md) for the complete configuration guide.

## Trusted Proxies

When running behind a reverse proxy (such as Nginx), configure the trusted proxy IPs so that Website Defender correctly identifies client IPs from the `X-Forwarded-For` header:

```yaml
trustedProxies:
  - "127.0.0.1"
  - "::1"
  - "10.0.0.0/8"
```

!!! info "Why Trusted Proxies Matter"
    Without trusted proxy configuration, rate limiting, IP blacklisting, and access logging may use the proxy's IP instead of the real client IP. Always include the IP addresses of your reverse proxy servers.

## Graceful Shutdown

Website Defender handles `SIGINT` and `SIGTERM` signals for graceful shutdown. When a shutdown signal is received:

1. The server stops accepting new connections
2. In-flight requests are allowed to complete
3. Database connections are closed cleanly

This makes it safe to use with process managers like `systemd`, `supervisord`, or container orchestrators.

## Running as a System Service

Example `systemd` unit file:

```ini
[Unit]
Description=Website Defender WAF
After=network.target

[Service]
Type=simple
ExecStart=/opt/defender/app
WorkingDirectory=/opt/defender
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

!!! tip "Working Directory"
    If using SQLite with the default path (`./data/app.db`), ensure the `WorkingDirectory` is set correctly so the database file is created in the expected location.
