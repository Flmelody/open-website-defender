# Nginx Setup

Website Defender integrates with Nginx via the `auth_request` module. This page provides the complete Nginx configuration for using Website Defender as an authentication provider.

## How auth_request Works

The Nginx `auth_request` module makes an internal subrequest to Website Defender's `/auth` endpoint for every incoming request. Based on the response:

- **200 OK**: Nginx proxies the request to the upstream application
- **401 Unauthorized**: The user is redirected to the guard (login) page
- **403 Forbidden**: The request is denied

## Full Configuration Example

```nginx
server {
    listen 80;
    server_name gitea.example.com;

    # All requests require authentication via Website Defender
    location / {
        auth_request /auth;

        # On 401, redirect to the guard page for login
        error_page 401 = @login_redirect;

        # Proxy to the actual application
        proxy_pass http://gitea-backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Internal auth subrequest to Website Defender
    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;

        # Pass the original domain for scope checking
        proxy_set_header X-Forwarded-Host $host;

        # Pass the real client IP for IP list checks and rate limiting
        proxy_set_header X-Forwarded-For $remote_addr;

        # Do not send the request body to the auth endpoint
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }

    # Redirect to guard page when authentication is required
    location @login_redirect {
        return 302 http://defender.example.com:9999/wall/guard/;
    }
}
```

## Configuration Explained

### The `/auth` Location Block

```nginx
location = /auth {
    internal;
    proxy_pass http://127.0.0.1:9999/wall/auth;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
}
```

| Directive | Purpose |
|-----------|---------|
| `internal` | Ensures this location can only be accessed via internal subrequests, not directly by clients |
| `proxy_pass` | Forwards the auth check to Website Defender's `/auth` endpoint |
| `X-Forwarded-Host` | Passes the original requested domain so Defender can perform [domain scope](../features/domain-scopes.md) checks |
| `X-Forwarded-For` | Passes the real client IP for [IP list](../features/ip-lists.md) checks and [rate limiting](../features/rate-limiting.md) |
| `proxy_pass_request_body off` | The auth check does not need the request body -- this improves performance |
| `Content-Length ""` | Required when disabling request body forwarding |

!!! warning "X-Forwarded-Host is Required for Domain Scopes"
    If you use [domain scopes](../features/domain-scopes.md), the `proxy_set_header X-Forwarded-Host $host;` directive is essential. Without it, Defender cannot determine which domain the user is trying to access.

### The auth_request Directive

```nginx
location / {
    auth_request /auth;
    ...
}
```

This tells Nginx to check with Website Defender before serving any content. The auth subrequest carries the original request's headers (including cookies), allowing Defender to validate the user's session.

## Multiple Protected Applications

To protect multiple applications behind the same Defender instance, create a server block for each:

```nginx
# Gitea
server {
    server_name gitea.example.com;

    location / {
        auth_request /auth;
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

# Jenkins
server {
    server_name jenkins.example.com;

    location / {
        auth_request /auth;
        proxy_pass http://jenkins-backend;
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

!!! tip "Use Domain Scopes"
    When protecting multiple applications, use [domain scopes](../features/domain-scopes.md) to control which users can access which services. For example, you can restrict a developer to `gitea.example.com` while giving an admin access to `*.example.com`.

## Defender Admin and Guard Pages

The Defender admin dashboard and guard page are served directly by the Go backend. You may optionally proxy them through Nginx as well:

```nginx
# Website Defender admin and guard
server {
    listen 80;
    server_name defender.example.com;

    location / {
        proxy_pass http://127.0.0.1:9999;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

!!! note "Do Not auth_request the Defender Itself"
    The Defender's own admin and guard pages should **not** be behind `auth_request`, as this would create a circular dependency. The admin dashboard has its own built-in authentication.
