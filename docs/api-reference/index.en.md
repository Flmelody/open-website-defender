# API Reference

All API routes are prefixed with the configurable `ROOT_PATH` (default: `/wall`). For example, the login endpoint is accessible at `/wall/login`.

## Authentication

Protected endpoints (marked **Yes** in the Auth column) require one of the following:

- `Defender-Authorization` header with a valid JWT token
- `flmelody.token` cookie with a valid JWT token

Obtain a token by calling `POST /login`.

## Endpoints

### Public Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/login` | User authentication. Returns a JWT token. | No |
| `GET` | `/auth` | Verify credentials (IP lists + token + scope check). Used by Nginx `auth_request`. | No |
| `GET` | `/health` | Health check endpoint. | No |

### Dashboard

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/dashboard/stats` | Dashboard statistics (request counts, uptime, etc.) | Yes |

### User Management

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/users` | List all users | Yes |
| `POST` | `/users` | Create a new user | Yes |
| `PUT` | `/users/:id` | Update a user | Yes |
| `DELETE` | `/users/:id` | Delete a user | Yes |

### IP Blacklist

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/ip-black-list` | List all blacklist entries | Yes |
| `POST` | `/ip-black-list` | Add an IP to the blacklist | Yes |
| `DELETE` | `/ip-black-list/:id` | Remove a blacklist entry | Yes |

### IP Whitelist

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/ip-white-list` | List all whitelist entries | Yes |
| `POST` | `/ip-white-list` | Add an IP to the whitelist | Yes |
| `DELETE` | `/ip-white-list/:id` | Remove a whitelist entry | Yes |

### WAF Rules

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/waf-rules` | List all WAF rules | Yes |
| `POST` | `/waf-rules` | Create a custom WAF rule | Yes |
| `PUT` | `/waf-rules/:id` | Update a WAF rule | Yes |
| `DELETE` | `/waf-rules/:id` | Delete a WAF rule | Yes |

### Access Logs

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/access-logs` | Query access logs with filters | Yes |
| `GET` | `/access-logs/stats` | Aggregated access log statistics | Yes |

### Authorized Domains

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/authorized-domains` | List authorized domains (paginated, or `?all=true` for all) | Yes |
| `POST` | `/authorized-domains` | Register a new authorized domain | Yes |
| `DELETE` | `/authorized-domains/:id` | Remove an authorized domain (cascades to whitelist and user scopes) | Yes |

### Geo-Blocking

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/geo-block-rules` | List all blocked country codes | Yes |
| `POST` | `/geo-block-rules` | Add a country code to the block list | Yes |
| `DELETE` | `/geo-block-rules/:id` | Remove a country code | Yes |

### Licenses

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/licenses` | List all licenses | Yes |
| `POST` | `/licenses` | Create a new license token | Yes |
| `DELETE` | `/licenses/:id` | Delete a license | Yes |

### System

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/system/settings` | Get current system settings | Yes |
| `PUT` | `/system/settings` | Update system settings | Yes |
| `POST` | `/system/reload` | Reload configuration and clear caches | Yes |

## Auth Endpoint Details

The `GET /auth` endpoint is the core of Website Defender's Nginx integration. It is called by Nginx's `auth_request` directive for every incoming request.

**Request headers inspected:**

| Header | Purpose |
|--------|---------|
| `X-Forwarded-For` | Client IP address (from trusted proxy) |
| `X-Forwarded-Host` | Original requested domain (for scope checking) |
| `Defender-Authorization` | JWT token |
| `Defender-Git-Token` | Git token (`username:token` format) |
| `Defender-License` | License token |
| `Cookie: flmelody.token` | JWT token via cookie |

**Response codes:**

| Code | Meaning |
|------|---------|
| `200` | Access granted |
| `401` | Authentication required (redirect to guard page) |
| `403` | Access denied (blacklisted, scope mismatch, etc.) |

For the auth verification flow, see [Authentication](../features/authentication.md).
