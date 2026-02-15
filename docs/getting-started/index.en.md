# Getting Started

This guide walks you through building and running Website Defender from source.

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| **Go** | 1.25+ | Backend compilation |
| **Node.js** | 20+ | Frontend build toolchain |
| **Nginx** | Any recent | Must include the `auth_request` module |

!!! info "Nginx auth_request Module"
    The `auth_request` module is included by default in most Nginx packages. You can verify by running `nginx -V` and checking for `--with-http_auth_request_module`.

## Build

The project includes a build script that compiles both the Vue 3 frontends and the Go backend into a single binary.

```bash
# 1. Clone the repository
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender

# 2. Build the project
./scripts/build.sh
```

!!! tip "Custom Build Configuration"
    You can customize build-time settings via environment variables or by editing `scripts/build.sh`. See [Environment Variables](../configuration/environment-variables.md) for available options.

The build script will:

1. Build the **guard** frontend (`ui/guard`) with Vite
2. Build the **admin** frontend (`ui/admin`) with Vite
3. Compile the Go backend with embedded frontend assets
4. Output a single binary named `app` in the project root

## Run

After building, start the application:

```bash
./app
```

The application starts with default configuration and is immediately ready to use.

| Setting | Default Value |
|---------|--------------|
| **Admin URL** | `http://localhost:9999/wall/admin/` |
| **Default Username** | `defender` |
| **Default Password** | `defender` |

!!! warning "Change Default Credentials"
    The default username and password are both `defender`. **Change these immediately** after first login, especially in production environments. Default credentials are a common attack vector.

## Next Steps

- **[Configure Nginx](../deployment/nginx-setup.md)** to use Website Defender as an auth provider
- **[Review the configuration](../configuration/index.md)** to customize runtime settings
- **[Set up the WAF](../features/waf.md)** and review built-in filtering rules
- **[Manage users](../features/user-management.md)** and assign domain scopes
