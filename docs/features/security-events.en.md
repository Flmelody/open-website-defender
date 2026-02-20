# Security Events

Website Defender records security events when threats are detected, providing a centralized view of all automated security actions.

## Event Types

| Type | Description | Trigger |
|------|-------------|---------|
| `auto_ban` | IP automatically banned | [Threat detection](threat-detection.md) threshold exceeded |
| `brute_force` | Brute force attempt detected | Excessive failed login attempts |
| `scan_detected` | Path scanning detected | Excessive 404 responses from a single IP |
| `js_challenge_fail` | JS challenge verification failed | Invalid or tampered challenge response |

## Admin Dashboard

The Security Events page in the admin dashboard provides:

### Statistics Overview

- **Total events** -- cumulative count of all security events
- **Auto-bans (24h)** -- number of automatic IP bans in the past 24 hours
- **Top threat IPs** -- IPs with the highest threat scores
- **Event type breakdown** -- distribution of events by type

### Event List

A filterable, paginated list of all security events with:

- Event type
- Client IP address
- Detail/reason
- Timestamp

### Filters

- **By event type** -- filter by `auto_ban`, `brute_force`, `scan_detected`, or `js_challenge_fail`
- **By client IP** -- search events for a specific IP address
- **By time range** -- narrow results to a specific time period

## Data Retention

Security events older than **90 days** are automatically deleted. This cleanup runs periodically in the background.

## Event Buffering

For performance, security events are buffered in memory (up to 500 events) and flushed to the database in batches every 5 seconds. This prevents database thrashing during high-traffic attack scenarios.

## API Access

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/security-events` | List security events with pagination and filters | Yes |
| `GET` | `/security-events/stats` | Get security event statistics | Yes |

---

## Related Pages

- [Threat Detection](threat-detection.md) -- How threats are detected and scored
- [Webhook](webhook.md) -- Get notified when security events occur
- [IP Lists](ip-lists.md) -- View and manage auto-banned IPs
- [Access Logs](access-logs.md) -- Request-level logging
