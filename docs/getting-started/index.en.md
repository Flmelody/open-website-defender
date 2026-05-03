# Getting Started

This guide walks you through building and running Castellum from source.

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
git clone https://github.com/Flmelody/castellum.git
cd castellum

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
| **Default Username** | `castellum` |
| **Default Password** | auto-generated on first start; read from `./data/bootstrap-admin-credentials` (mode `0600`) |

!!! warning "Change Default Credentials"
    On first start the default password is randomly generated and written to `./data/bootstrap-admin-credentials` (mode `0600`). The startup log only prints the file path — open the file to read the password. After your first login, **rotate the credentials and delete the file**. To skip auto-generation, set `default-user.username` / `default-user.password` in `config/config.yaml`.

!!! warning "Upgrading from Open Website Defender (PostgreSQL / MySQL)"
    The fallback default database name was changed from `open_website_defender` to `castellum`. If your `config/config.yaml` does **not** set `database.name` and you are upgrading an existing PG/MySQL deployment, either set `database.name: open_website_defender` explicitly or rename the database to `castellum` before upgrading. SQLite deployments are not affected.

## Next Steps

- **[Configure Nginx](../deployment/nginx-setup.md)** to use Castellum as an auth provider
- **[Review the configuration](../configuration/index.md)** to customize runtime settings
- **[Set up the WAF](../features/waf.md)** and review built-in filtering rules
- **[Register authorized domains](../features/authorized-domains.md)** and **[manage users](../features/user-management.md)** with authorized domain assignment
