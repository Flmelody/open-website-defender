# Development

This guide covers building Website Defender from source and setting up a local development environment.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.25+ | Backend compilation |
| **Node.js** | 20+ | Frontend build and dev server |
| **npm** | Included with Node.js | Frontend dependency management |

## Project Structure

```
open-website-defender/
├── main.go                    # Application entry point
├── config/
│   └── config.yaml            # Runtime configuration
├── scripts/
│   └── build.sh               # Build script
├── internal/
│   ├── adapter/
│   │   ├── controller/http/   # Gin HTTP handlers, middleware, requests, responses
│   │   └── repository/        # GORM repository implementations
│   ├── domain/
│   │   ├── entity/            # Domain models (User, IpBlackList, IpWhiteList, etc.)
│   │   └── error/             # Domain-specific errors
│   ├── infrastructure/
│   │   ├── config/            # Configuration structs
│   │   ├── database/          # Database initialization (SQLite/PostgreSQL/MySQL via GORM)
│   │   └── logging/           # Zap logger setup
│   ├── pkg/                   # Utilities (JWT, crypto, HTTP helpers)
│   └── usecase/               # Business logic services with DTOs
│       ├── interface/         # Repository interfaces
│       ├── iplist/            # IP list services
│       └── user/              # Auth and user services
├── ui/
│   ├── admin/                 # Admin dashboard (Vue 3 + Element Plus + vue-i18n)
│   └── guard/                 # Guard/challenge page (Vue 3)
└── docs/                      # MkDocs documentation
```

## Build from Source

### Full Build (Frontend + Backend)

The build script compiles both Vue 3 frontends and the Go backend into a single binary:

```bash
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender
./scripts/build.sh
```

The resulting `app` binary has all frontend assets embedded via `go:embed`.

### Manual Build Steps

If you prefer to build each component separately:

=== "Backend Only"

    ```bash
    # Requires frontend dist/ folders to exist (for go:embed)
    go build -o open-website-defender main.go
    ```

    !!! warning "Frontend Assets Required"
        The Go binary embeds `ui/admin/dist` and `ui/guard/dist` via `//go:embed`. You must build both frontends before compiling the backend, or the build will fail.

=== "Admin Frontend"

    ```bash
    cd ui/admin
    npm install
    npm run build
    ```

    Output: `ui/admin/dist/`

=== "Guard Frontend"

    ```bash
    cd ui/guard
    npm install
    npm run build
    ```

    Output: `ui/guard/dist/`

## Development Server

### Backend

Run the Go backend directly:

```bash
go run main.go
```

By default, it listens on port `9999` and serves at `http://localhost:9999/wall/`.

To use a custom config file:

```bash
go run main.go -config ./config/config.yaml
```

### Frontend Dev Servers

Each Vue app has its own Vite development server with hot module replacement:

=== "Admin Dashboard"

    ```bash
    cd ui/admin
    npm install
    npm run dev
    ```

=== "Guard Page"

    ```bash
    cd ui/guard
    npm install
    npm run dev
    ```

!!! tip "Frontend Development"
    During frontend development, the Vite dev server proxies API requests to the Go backend. Make sure the backend is running before starting the frontend dev server.

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/usecase/...
```

### Frontend

```bash
# Type checking (admin)
cd ui/admin
npm run type-check

# Lint and fix (admin)
npm run lint
```

## Key Patterns

### Singleton Services

Services use the `sync.Once` pattern for thread-safe initialization:

```go
var (
    authService     *AuthService
    authServiceOnce sync.Once
)

func GetAuthService() *AuthService {
    authServiceOnce.Do(func() {
        authService = &AuthService{...}
    })
    return authService
}
```

### Repository Interfaces

Repository interfaces are defined in `internal/usecase/interface/repository.go` and implemented in `internal/adapter/repository/`. This clean architecture pattern allows easy testing and database swapping.

### Authentication Flow

The auth middleware chain:

1. **IP Blacklist check** -- immediately deny blocked IPs
2. **IP Whitelist check** -- immediately allow trusted IPs
3. **JWT Token validation** -- check `Defender-Authorization` header or `flmelody.token` cookie
4. **Authorized Domain check** -- verify the user can access the requested domain
5. **Git Token validation** -- check `Defender-Git-Token` header
6. **License Token validation** -- check `Defender-License` header

### API Routes

All routes are prefixed with the configurable `ROOT_PATH` (default `/wall`) and protected routes use `AuthMiddleware`. See the [API Reference](../api-reference/index.md) for the complete endpoint list.
