# User Management

Website Defender provides full user management capabilities through the admin dashboard and API, including role-based access control, token generation, and domain scope assignment.

## User CRUD Operations

Administrators can create, view, edit, and delete users through the admin dashboard or the [API](../api-reference/index.md).

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/users` | List all users |
| `POST` | `/users` | Create a new user |
| `PUT` | `/users/:id` | Update a user |
| `DELETE` | `/users/:id` | Delete a user |

## Role-Based Access

Each user has an **admin privilege flag** that controls their access level:

| Role | Capabilities |
|------|-------------|
| **Admin** | Full access to all dashboard features, user management, and settings. Bypasses domain scope checks. |
| **Regular User** | Can authenticate and access protected services within their assigned domain scopes. |

!!! note "At Least One Admin"
    Ensure at least one admin user exists at all times. The default user created on first startup (`defender`) has admin privileges.

## Git Token Generation

Each user can have a Git token for machine access (CI/CD, scripts, automated tools):

- Tokens are **auto-generated** via the admin dashboard
- **One-click copy** for easy integration
- Token format: `username:token` (sent via the `Defender-Git-Token` header)
- Git tokens are subject to the user's [domain scope](domain-scopes.md) restrictions

!!! tip "Regenerating Git Tokens"
    You can regenerate a user's Git token at any time from the admin dashboard. The previous token is immediately invalidated.

## Domain Scopes

Each user can be assigned domain scopes that restrict which protected services they can access:

- Scopes are defined as comma-separated patterns (e.g., `gitea.com, *.internal.org`)
- Empty scopes grant unrestricted access
- Admin users bypass scope checks regardless of their scope configuration

For full details on how domain scopes work, see [Domain Scopes](domain-scopes.md).

## License Management

Website Defender also supports license tokens for API and machine access, managed separately from user accounts.

### Generating Licenses

- Generate new license tokens via the admin dashboard
- Tokens are **shown only once** at creation time -- copy and store them securely
- Tokens are stored as **SHA-256 hashes** in the database, so the original token cannot be recovered

### Managing Licenses

| Action | Description |
|--------|-------------|
| **Activate** | Enable a license token for use |
| **Revoke** | Disable a license token, immediately preventing access |
| **Delete** | Permanently remove a license entry |

!!! warning "Store Licenses Securely"
    License tokens are displayed only once when generated. If lost, the token cannot be recovered -- you will need to generate a new one.

Licenses can also be managed via the API:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/licenses` | List all licenses |
| `POST` | `/licenses` | Create a new license |
| `DELETE` | `/licenses/:id` | Delete a license |
