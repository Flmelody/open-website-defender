# IP Lists

Website Defender provides IP whitelist and blacklist management to control access at the network level. IP checks are the **first step** in the [auth verification flow](authentication.md), evaluated before any token-based authentication.

## IP Whitelist

The IP whitelist allows specific IP addresses or CIDR ranges to **bypass all authentication checks**. Whitelisted IPs are granted access immediately without any token validation.

Use cases:

- Internal network ranges that should always have access
- Monitoring systems and health check probes
- Trusted CI/CD infrastructure

### Supported Formats

| Format | Example | Description |
|--------|---------|-------------|
| Exact IP | `192.168.1.100` | Matches a single IP address |
| CIDR range | `192.168.1.0/24` | Matches all IPs in the subnet |
| IPv6 | `::1` | Supports IPv6 addresses |

!!! tip "Use CIDR for Internal Networks"
    Rather than whitelisting individual IPs, use CIDR notation to whitelist entire subnets. For example, `10.0.0.0/8` covers all private IPs in the `10.x.x.x` range.

## IP Blacklist

The IP blacklist blocks specific IP addresses or CIDR ranges **before any other processing**. Blacklisted IPs receive an immediate `403 Forbidden` response.

Use cases:

- Blocking known malicious IPs
- Blocking IP ranges associated with abuse
- Emergency blocking during active attacks

### Supported Formats

The same formats as the whitelist are supported: exact IP, CIDR range, and IPv6.

!!! warning "Blacklist Takes Priority"
    The blacklist is checked **before** the whitelist. If an IP appears in both lists, it will be blocked. Always review your blacklist entries to avoid accidentally blocking trusted IPs.

## Management

### Admin Dashboard

Both IP lists can be managed through the admin dashboard:

- Add new entries with optional descriptions
- View all current entries
- Delete individual entries

### API

IP lists can also be managed programmatically via the REST API:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/ip-black-list` | List all blacklist entries |
| `POST` | `/ip-black-list` | Add a blacklist entry |
| `DELETE` | `/ip-black-list/:id` | Remove a blacklist entry |
| `GET` | `/ip-white-list` | List all whitelist entries |
| `POST` | `/ip-white-list` | Add a whitelist entry |
| `DELETE` | `/ip-white-list/:id` | Remove a whitelist entry |

All API routes require authentication. See the [API Reference](../api-reference/index.md) for full details.

## Auth Flow Position

In the auth verification flow, IP checks are evaluated first:

```
IP Blacklist → IP Whitelist → JWT Token → Git Token → License Token → Deny
```

This means IP-level decisions are made before any token parsing or validation occurs, providing fast rejection of known-bad actors and fast acceptance of known-good infrastructure.
