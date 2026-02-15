# Access Logs & Analytics

Website Defender logs all incoming requests and provides a dashboard for analyzing traffic patterns and identifying blocked threats.

## Logged Fields

Each request is recorded with the following information:

| Field | Description |
|-------|-------------|
| **Client IP** | The IP address of the requester (respects `X-Forwarded-For` from trusted proxies) |
| **Method** | HTTP method (`GET`, `POST`, `PUT`, `DELETE`, etc.) |
| **Path** | The requested URL path |
| **Status Code** | HTTP response status code |
| **Latency** | Request processing time |
| **User-Agent** | The client's User-Agent header |
| **Action** | Whether the request was `allowed` or `blocked` |

## Dashboard Analytics

The admin dashboard provides real-time analytics based on the access logs:

- **Total request count** -- overall number of requests processed
- **Blocked request count** -- number of requests denied by any security layer
- **Top 10 blocked IPs** -- most frequently blocked IP addresses
- **Filtering** -- search and filter logs by:
    - IP address
    - Action (allowed / blocked)
    - Status code
    - Time range

!!! tip "Identify Attack Patterns"
    Use the Top 10 blocked IPs view to identify persistent attackers. You can then add those IPs to the [IP Blacklist](ip-lists.md) for permanent blocking.

## API Access

Access logs and statistics are also available via the API:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/access-logs` | Query access logs with filters |
| `GET` | `/access-logs/stats` | Get aggregated access log statistics |

Both endpoints require authentication. See the [API Reference](../api-reference/index.md) for full details.
