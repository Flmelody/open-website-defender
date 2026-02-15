# Docker Deployment

Website Defender provides a multi-stage Dockerfile for building a minimal container image with all frontend assets and the Go binary.

## Quick Start

### Using Docker Compose (Recommended)

```bash
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender

# Copy and edit environment variables
cp .env.example .env
vim .env

# Edit runtime config as needed
vim config/config.yaml

# Build and start
docker compose up -d
```

The service will be available at `http://localhost:9999`.

Docker Compose automatically reads the `.env` file in the project root and passes the variables as build arguments to the Dockerfile. This ensures all configuration stays in one place.

### Using Docker Directly

```bash
# Build the image (uses Dockerfile ARG defaults)
docker build -t defender .

# Or with custom build arguments
docker build \
  --build-arg BACKEND_HOST=https://example.com/wall \
  -t defender .

# Run the container (works without any volume mounts)
docker run -d \
  --name defender \
  -p 9999:9999 \
  defender
```

## Build Arguments

The Dockerfile accepts build arguments to customize the frontend at build time. These values are baked into the frontend assets via Vite and into the Go binary via ldflags.

| Argument | Default | Description |
|----------|---------|-------------|
| `BACKEND_HOST` | `http://localhost:9999/wall` | Full URL the browser uses to reach the backend API |
| `ROOT_PATH` | `/wall` | API route prefix |
| `ADMIN_PATH` | `/admin` | Admin dashboard path |
| `GUARD_PATH` | `/guard` | Guard/challenge page path |

!!! warning "BACKEND_HOST Must Match Your Actual Access URL"
    `BACKEND_HOST` is embedded into the frontend JavaScript at build time. It must match the URL that users will actually use to access the service (including the correct host and port).

    For example, if you map the container to port 8080 (`-p 8080:9999`), you **must** set:

    ```
    BACKEND_HOST=http://localhost:8080/wall
    ```

    Otherwise the frontend will still try to reach port 9999, which will fail.

### Configuration Flow

All build arguments share the same variable names as the `.env` file used by `scripts/build.sh`:

```
.env  ──▶  docker-compose.yml (${VAR:-default})  ──▶  Dockerfile ARG
                                                         ▼
scripts/build.sh  ◀── .env                        Vite env + Go ldflags
```

- **`docker compose build`**: reads `.env` automatically, passes values as build args, overriding Dockerfile ARG defaults.
- **`docker build`** (standalone): uses Dockerfile ARG defaults directly. Pass `--build-arg` to override.

## Port Configuration

The `PORT` variable from `.env` drives the entire port chain — the application listening port inside the container, the Docker port mapping, and the `EXPOSE` directive:

```
.env (PORT=8080)
  ├──▶ build arg   ──▶ Dockerfile EXPOSE 8080
  ├──▶ ports       ──▶ "8080:8080"
  └──▶ environment ──▶ Go app listens on :8080
```

To change the port, set `PORT` and `BACKEND_HOST` together in `.env`:

```bash
# .env
PORT=8080
BACKEND_HOST=http://localhost:8080/wall
```

Then rebuild:

```bash
docker compose up -d --build
```

!!! note
    `BACKEND_HOST` is baked into the frontend at build time. When changing `PORT`, always update `BACKEND_HOST` to match and rebuild the image.

## Volumes

The container works out of the box **without any volume mounts** — default configuration and an SQLite database are embedded in the image. Volume mounts are optional and only needed when you want to persist data or override the default configuration.

| Path | Purpose | Required |
|------|---------|----------|
| `/app/data` | SQLite database (`app.db`) and other persistent data | No (recommended for production) |
| `/app/config` | Runtime configuration (`config.yaml`) | No (image includes defaults) |

```bash
# Minimal: no mounts, uses built-in defaults
docker run -d -p 9999:9999 defender

# Production: mount data for persistence and config for customization
docker run -d \
  -p 9999:9999 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config:/app/config \
  defender
```

!!! warning "Data Persistence"
    Without mounting `/app/data`, the SQLite database lives inside the container and will be **lost** when the container is removed. For production use, always mount this path.

## Configuration

The image ships with a default `config/config.yaml`. To customize it, mount your own config file:

```bash
docker run -d \
  -v $(pwd)/config:/app/config \
  -p 9999:9999 \
  defender
```

The configuration file format is the same as for bare-metal deployments — see [Configuration](../configuration/index.md).

## Using PostgreSQL

For production deployments, you can use PostgreSQL instead of SQLite.

1. Uncomment the `postgres` service in `docker-compose.yml`
2. Update `config/config.yaml`:

```yaml
database:
  driver: postgres
  host: postgres
  port: 5432
  name: open_website_defender
  user: postgres
  password: changeme
```

3. Start both services:

```bash
docker compose up -d
```

## Docker Networking and Client IP Detection

!!! danger "IP-Based Features Require a Reverse Proxy or Host Network"
    When using Docker's default bridge network with port mapping (`-p 9999:9999`), **all requests appear to come from the Docker gateway IP** (e.g. `172.19.0.1`) instead of the real client IP. This is because Docker port mapping is Layer 4 NAT — it does not set any HTTP headers.

    This affects **all IP-based features**:

    | Feature | Impact |
    |---------|--------|
    | IP Blacklist | Blocking an external IP has no effect |
    | IP Whitelist | Allowing an external IP has no effect |
    | Rate Limiting | All users share one IP's quota, causing false triggers |
    | Access Logs | All logs show the gateway IP, not the real client |
    | Geo-IP Blocking | Cannot determine real client location |

    Blacklisting the gateway IP (e.g. `172.19.0.1`) will **block all requests entirely**.

### Solution 1: Nginx Reverse Proxy (Recommended)

Place Nginx in front of the container. Nginx sets the `X-Forwarded-For` header with the real client IP:

```
Client (1.2.3.4) ──▶ Nginx ──▶ Docker Container
                       │
                       └── X-Forwarded-For: 1.2.3.4
```

Then add the Nginx/Docker gateway IP to `trustedProxies` in `config.yaml`:

```yaml
trustedProxies:
  - "172.16.0.0/12"   # Docker network range
  - "127.0.0.1"
```

See [Nginx Setup](nginx-setup.md) for the complete configuration.

### Solution 2: Host Network Mode

Use `network_mode: host` so the container shares the host's network stack and sees real client IPs directly:

```yaml
# docker-compose.yml
services:
  defender:
    # ...
    network_mode: host
```

No `trustedProxies` configuration is needed in this mode. Note that `ports` mapping is ignored with host networking — the application binds directly to the host port.

## Production Tips

!!! tip "Production Checklist"
    - Set a stable `security.jwt-secret` in `config.yaml`
    - Change the default credentials (`defender/defender`)
    - Use PostgreSQL or MySQL for better concurrency
    - Configure `trustedProxies` to include your reverse proxy IPs
    - Set explicit `security.cors.allowed-origins`
    - Use a named Docker volume or bind mount for `/app/data`
    - Set `BACKEND_HOST` to the actual production URL before building

### Running Behind Nginx

When running the Docker container behind Nginx, add the Docker network gateway or Nginx host IP to `trustedProxies`:

```yaml
trustedProxies:
  - "172.17.0.1"   # Default Docker bridge gateway
  - "127.0.0.1"
```

See [Nginx Setup](nginx-setup.md) for the full reverse proxy configuration.

### Health Checks

Add a health check to your Docker Compose configuration:

```yaml
services:
  defender:
    # ...
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9999/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

### Resource Limits

```yaml
services:
  defender:
    # ...
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "1.0"
```
