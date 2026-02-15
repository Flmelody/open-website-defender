# Domain Scopes

Domain scopes enable **multi-tenant access control**, allowing different users to access different protected services behind the same Website Defender instance.

## How It Works

1. When a request hits the `/auth` endpoint, Defender reads the target domain from the `X-Forwarded-Host` header (with fallback to the `Host` header)
2. After successful token or Git token authentication, the user's configured scopes are checked against the requested domain
3. If the domain does not match any of the user's scope patterns, a `403 Forbidden` response is returned

!!! info "Domain Source"
    The domain is extracted from the `X-Forwarded-Host` header first, falling back to the `Host` header if not present. Make sure your Nginx configuration passes this header -- see [Nginx Setup](../deployment/nginx-setup.md).

## Scope Patterns

Scopes are defined as comma-separated patterns on each user. The following pattern types are supported:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `gitea.com` | `gitea.com` | `gitlab.com`, `sub.gitea.com` |
| `*.example.com` | `app.example.com`, `dev.example.com` | `example.com` |
| `gitea.com, *.internal.org` | `gitea.com`, `app.internal.org` | `gitlab.com` |
| *(empty)* | Everything (unrestricted) | - |

## Rules

- **Empty scopes** grant unrestricted access -- this maintains backward compatibility with existing users who have no scopes configured
- **Admin users** always bypass scope checks, regardless of their scopes value
- Matching is **case-insensitive** (`Gitea.COM` matches scope `gitea.com`)
- **Ports are stripped** before matching (`gitea.com:3000` matches scope `gitea.com`)

!!! tip "Start with Empty Scopes"
    When migrating an existing deployment to use domain scopes, all existing users will continue to have unrestricted access since their scopes are empty. You can then progressively restrict users as needed.

## Nginx Configuration

To pass the domain information to Defender for scope checking, configure Nginx to forward the `Host` header via `X-Forwarded-Host`:

```nginx
server {
    server_name gitea.example.com;

    location / {
        auth_request /auth;

        # Pass the original host to Defender for scope checking
        proxy_pass http://gitea-backend;
    }

    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
}
```

!!! warning "X-Forwarded-Host Required"
    Without the `proxy_set_header X-Forwarded-Host $host;` directive, Defender will fall back to the `Host` header, which in an `auth_request` subrequest context may not reflect the original requested domain.

For the complete Nginx configuration reference, see [Nginx Setup](../deployment/nginx-setup.md).

## Example Scenario

Consider an organization with the following protected services:

| Service | Domain |
|---------|--------|
| Gitea | `gitea.internal.org` |
| Jenkins | `jenkins.internal.org` |
| Grafana | `grafana.internal.org` |

You could configure users as follows:

| User | Scopes | Access |
|------|--------|--------|
| `alice` | `*.internal.org` | All services |
| `bob` | `gitea.internal.org, jenkins.internal.org` | Gitea and Jenkins only |
| `charlie` | `grafana.internal.org` | Grafana only |
| `admin` | *(any)* | All services (admin bypass) |
