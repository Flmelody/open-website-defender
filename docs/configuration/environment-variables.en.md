# Environment Variables

Website Defender uses a small set of environment variables at **build time** to configure frontend path prefixes. Host, domain, and port settings are **runtime configuration** managed through `config.yaml` or OS environment variables, so a single pre-built binary works across different deployments.

!!! info "Build Time vs Runtime"
    Only path variables (`ROOT_PATH`, `ADMIN_PATH`, `GUARD_PATH`) are injected during the build. They are compiled into the binary via Vite env vars and Go ldflags, and cannot be changed after building. Host, domain, and port settings are read at startup -- see [Runtime Configuration](#runtime-configuration) below.

## Build-Time Variables (.env)

These variables must be set **before** running `scripts/build.sh`. They determine the URL path structure baked into both the Vue frontends and the Go binary.

| Variable | Default | Description |
|----------|---------|-------------|
| `ROOT_PATH` | `/wall` | The root path context for all routes |
| `ADMIN_PATH` | `/admin` | The path for the admin dashboard |
| `GUARD_PATH` | `/guard` | The path for the guard (challenge/login) page |

### Via Environment Variables

Set the variables before running the build script:

```bash
export ROOT_PATH="/wall"
export ADMIN_PATH="/admin"
export GUARD_PATH="/guard"

./scripts/build.sh
```

### Via .env File

Create a `.env` file in the project root. The build script automatically loads it:

```bash
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
```

Then build normally:

```bash
./scripts/build.sh
```

## Runtime Configuration

Host, domain, and port settings are resolved at **startup**, not at build time. This means a single compiled binary can be deployed to different environments by changing only the config file or OS environment variables.

### `wall:` Section (config.yaml)

The `wall:` section in `config.yaml` controls how the frontend reaches the backend and how auth cookies are scoped:

```yaml
wall:
  # API base URL for the frontend (default: same-origin, using root-path)
  backend-host: "https://defender.example.com/wall"
  # Cookie domain for SSO across subdomains (e.g. ".example.com")
  guard-domain: ".example.com"
```

| Key | Default | Description |
|-----|---------|-------------|
| `wall.backend-host` | *(empty -- same-origin)* | The backend API URL used by the frontend. Leave empty when the frontend and backend share the same origin. Set it for cross-origin or reverse-proxy setups. |
| `wall.guard-domain` | *(empty)* | Cookie domain for the `flmelody.token` auth cookie. When set (e.g. `.example.com`), the cookie is shared across all subdomains, enabling single sign-on. |

These can also be set via OS environment variables `BACKEND_HOST` and `GUARD_DOMAIN`, which take the same precedence as other Viper-bound env vars.

### Port

The listening port is configured in the `server:` section of `config.yaml`, or via the `PORT` OS environment variable:

```yaml
server:
  port: 9999
```

```bash
# Or as an environment variable at runtime
PORT=8080 ./app
```

| Source | Default | Description |
|--------|---------|-------------|
| `server.port` (config.yaml) | `9999` | The port the server listens on |
| `PORT` (env var) | `9999` | OS environment variable override for the port |

!!! info "Runtime Config Injection"
    The backend injects a `window.__APP_CONFIG__` object into the frontend's `index.html` at startup. This object carries the current `backend-host`, `guard-domain`, and path settings to the browser. Because injection happens on every request, pre-built binaries work across different deployments without rebuilding the frontend.

## How It Works

The build script (`scripts/build.sh`) performs the following:

1. Reads path variables from the environment or `.env` file (with defaults if not set)
2. Exports them as `VITE_*` variables for the Vue 3 frontend builds
3. Passes them as Go `ldflags` to embed in the backend binary

```bash
# Frontend receives paths as VITE_ prefixed variables
export VITE_ROOT_PATH=$ROOT_PATH
export VITE_ADMIN_PATH=$ADMIN_PATH
export VITE_GUARD_PATH=$GUARD_PATH

# Backend receives paths via ldflags
go build -ldflags "\
  -X 'main.RootPath=$ROOT_PATH' \
  -X 'main.AdminPath=$ADMIN_PATH' \
  -X 'main.GuardPath=$GUARD_PATH' \
" -o app main.go
```

Host and domain values are **not** passed through the build. They are read from `config.yaml` at startup and injected into the served HTML via `window.__APP_CONFIG__`.

## URL Structure

With the default configuration, the application serves at:

| Resource | URL |
|----------|-----|
| Admin Dashboard | `http://localhost:9999/wall/admin/` |
| Guard Page | `http://localhost:9999/wall/guard/` |
| API Root | `http://localhost:9999/wall/` |
| Auth Endpoint | `http://localhost:9999/wall/auth` |
| Login Endpoint | `http://localhost:9999/wall/login` |
| Admin Login Endpoint | `http://localhost:9999/wall/admin-login` |

!!! tip "Custom Path Example"
    If you set `ROOT_PATH=/defender` and `ADMIN_PATH=/dashboard`, the admin URL becomes `http://localhost:9999/defender/dashboard/`.
