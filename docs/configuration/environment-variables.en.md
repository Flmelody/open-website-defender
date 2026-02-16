# Environment Variables

Website Defender uses environment variables at **build time** to configure frontend paths and backend host settings. These values are embedded into the compiled binary during the build process.

!!! info "Build Time vs Runtime"
    These environment variables are injected during the build (`scripts/build.sh`) and compiled into the binary. They cannot be changed after building. For runtime configuration, see [Configuration](index.md).

## Available Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BACKEND_HOST` | `http://localhost:9999/wall` | The backend API host URL used by the frontend |
| `ROOT_PATH` | `/wall` | The root path context for all routes |
| `ADMIN_PATH` | `/admin` | The path for the admin dashboard |
| `GUARD_PATH` | `/guard` | The path for the guard (challenge/login) page |

## Usage

### Via Environment Variables

Set the variables before running the build script:

```bash
export BACKEND_HOST="https://defender.example.com/wall"
export ROOT_PATH="/wall"
export ADMIN_PATH="/admin"
export GUARD_PATH="/guard"

./scripts/build.sh
```

### Via .env File

Create a `.env` file in the project root. The build script automatically loads it:

```bash
BACKEND_HOST=https://defender.example.com/wall
ROOT_PATH=/wall
ADMIN_PATH=/admin
GUARD_PATH=/guard
```

Then build normally:

```bash
./scripts/build.sh
```

## How It Works

The build script (`scripts/build.sh`) performs the following:

1. Reads environment variables (with defaults if not set)
2. Exports them as `VITE_*` variables for the Vue 3 frontend builds
3. Passes them as Go `ldflags` to embed in the backend binary

```bash
# Frontend receives them as VITE_ prefixed variables
export VITE_BACKEND_HOST=$BACKEND_HOST
export VITE_ROOT_PATH=$ROOT_PATH

# Backend receives them via ldflags
go build -ldflags "-X 'main.BackendHost=$BACKEND_HOST' ..." -o app main.go
```

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
