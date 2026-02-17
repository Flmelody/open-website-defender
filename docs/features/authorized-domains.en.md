# Authorized Domains

Authorized Domains is a centralized registry of all protected domains managed by Website Defender. It serves as the **single source of truth** for domain names used across the system -- in IP whitelist bindings and user access control.

## Purpose

Without Authorized Domains, users must manually type domain names when configuring IP whitelist entries or user access settings, which is error-prone and inconsistent. Authorized Domains solves this by:

- Providing a **central place** to register all protected domains
- Populating **dropdown selectors** in the IP whitelist and user management forms

- Enabling **multi-tenant access control**, allowing different users to access different protected services behind the same Defender instance

## Management

### Admin Dashboard

Manage authorized domains through the admin dashboard:

1. Navigate to the **Authorized Domains** page
2. Add domains by entering the domain name (e.g., `app.example.com`)
3. View all registered domains with creation timestamps
4. Delete domains when they are no longer needed

### API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/authorized-domains` | List authorized domains (paginated) |
| `GET` | `/authorized-domains?all=true` | List all authorized domains (for dropdowns) |
| `POST` | `/authorized-domains` | Register a new authorized domain |
| `DELETE` | `/authorized-domains/:id` | Remove an authorized domain |

!!! note "Duplicate Prevention"
    Domain names must be unique. Attempting to register a domain that already exists will return a `409 Conflict` response.

## Access Control

Authorized domains enable **multi-tenant access control**. Each user can be assigned specific authorized domains, restricting which protected services they can access.

### How It Works

1. When a request hits the `/auth` endpoint, Defender reads the target domain from the `X-Forwarded-Host` header (with fallback to the `Host` header)
2. After successful token or Git token authentication, the user's configured authorized domains are checked against the requested domain
3. If the domain does not match any of the user's authorized domain patterns, a `403 Forbidden` response is returned

!!! info "Domain Source"
    The domain is extracted from the `X-Forwarded-Host` header first, falling back to the `Host` header if not present. Make sure your Nginx configuration passes this header -- see [Nginx Setup](../deployment/nginx-setup.md).

### Domain Patterns

Authorized domains are defined as comma-separated patterns on each user, typically selected from the registry. The following pattern types are supported:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `app.example.com` | `app.example.com` | `other.example.com`, `sub.app.example.com` |
| `*.example.com` | `app.example.com`, `dev.example.com` | `example.com` |
| `app.example.com, *.internal.org` | `app.example.com`, `svc.internal.org` | `other.example.com` |
| *(empty)* | Everything (unrestricted) | - |

### Rules

- **Empty authorized domains** grant unrestricted access -- this maintains backward compatibility with existing users who have no domains configured
- **Admin users** always bypass authorized domain checks, regardless of their configuration
- Matching is **case-insensitive** (`App.Example.COM` matches `app.example.com`)
- **Ports are stripped** before matching (`app.example.com:3000` matches `app.example.com`)

!!! tip "Start with Empty Domains"
    When migrating an existing deployment to use authorized domain access control, all existing users will continue to have unrestricted access since their authorized domains are empty. You can then progressively restrict users as needed.

### Nginx Configuration

To pass the domain information to Defender for access control, configure Nginx to forward the `Host` header via `X-Forwarded-Host`:

```nginx
server {
    server_name app.example.com;

    location / {
        auth_request /auth;

        # Pass the original host to Defender for authorized domain checking
        proxy_pass http://app-backend;
    }

    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Original-URI $request_uri;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
}
```

!!! warning "X-Forwarded-Host Required"
    Without the `proxy_set_header X-Forwarded-Host $host;` directive, Defender will fall back to the `Host` header, which in an `auth_request` subrequest context may not reflect the original requested domain.

For the complete Nginx configuration reference, see [Nginx Setup](../deployment/nginx-setup.md).

### Example Scenario

Consider an organization with the following protected services:

| Service | Domain |
|---------|--------|
| App 1 | `app1.internal.org` |
| App 2 | `app2.internal.org` |
| App 3 | `app3.internal.org` |

You could configure users as follows:

| User | Authorized Domains | Access |
|------|-------------------|--------|
| `alice` | `*.internal.org` | All services |
| `bob` | `app1.internal.org, app2.internal.org` | App 1 and App 2 only |
| `charlie` | `app3.internal.org` | App 3 only |
| `admin` | *(any)* | All services (admin bypass) |

## Integration with Other Features

### IP Whitelist

When adding an IP whitelist entry, the **Authorized Domain** field provides a dropdown populated from the authorized domains registry. You can also type a custom value if needed.

See [IP Lists](ip-lists.md) for details.

### User Management

When configuring user access, the **Authorized Domains** field provides a multi-select dropdown populated from the registry. Users can select multiple domains or type custom patterns (e.g., `*.example.com`).

See [User Management](user-management.md) for details.

---

## Related Pages

- [IP Lists](ip-lists.md) -- IP whitelist domain binding
- [User Management](user-management.md) -- User authorized domain assignment
- [Authentication](authentication.md) -- Full auth verification flow
- [Nginx Setup](../deployment/nginx-setup.md) -- Complete Nginx configuration guide
- [API Reference](../api-reference/index.md) -- Full API documentation
