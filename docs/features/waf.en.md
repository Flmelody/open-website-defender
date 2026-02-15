# Web Application Firewall (WAF)

Website Defender includes a regex-based Web Application Firewall that inspects incoming requests for common attack patterns including SQL injection, cross-site scripting (XSS), and path traversal.

## How It Works

The WAF inspects the following parts of each request:

- **URL path**
- **Query string**
- **User-Agent header**
- **Request body** (up to 10KB)

Each WAF rule supports one of two actions:

| Action | Behavior |
|--------|----------|
| **block** | Returns `403 Forbidden` and logs the request |
| **log** | Allows the request but records it in the access log |

!!! tip "Start with Log Mode"
    When deploying new custom rules, consider using `log` mode first to observe matches without blocking legitimate traffic. Switch to `block` once you are confident the rule has no false positives.

## Built-in Rules

Website Defender ships with **9 built-in rules** covering the most common web attack categories:

### SQL Injection

| Rule | Description |
|------|-------------|
| Union Select | Detects `UNION SELECT` based attacks |
| Common Patterns | Detects `; DROP`, `; ALTER`, `; DELETE`, and similar statements |
| Boolean Injection | Detects `' OR 1=1` style authentication bypass |
| Comment Injection | Detects `' --` and `/* */` comment abuse |

### Cross-Site Scripting (XSS)

| Rule | Description |
|------|-------------|
| Script Tag | Detects `<script>` tag injection |
| Event Handler | Detects `onerror=`, `onclick=`, and other inline event handlers |
| JavaScript Protocol | Detects `javascript:` and `vbscript:` protocol handlers |

### Path Traversal

| Rule | Description |
|------|-------------|
| Dot Dot Slash | Detects `../`, `..\`, and URL-encoded variants |
| Sensitive Files | Detects access attempts to `/etc/passwd`, `/proc/self`, and similar system files |

## Custom Rules

In addition to the built-in rules, you can create custom WAF rules via the admin dashboard or the [API](../api-reference/index.md).

Custom rules follow the same structure:

- A **regex pattern** to match against request components
- An **action** (`block` or `log`)
- An optional **description**

!!! note "Rule Management"
    WAF rules (both built-in and custom) can be managed through the admin dashboard under the WAF Rules section, or programmatically via the `/waf-rules` API endpoints.

## Configuration

The WAF can be enabled or disabled globally in `config/config.yaml`:

```yaml
request-filtering:
  enabled: true
```

When disabled, no request inspection is performed and all WAF rules are skipped.

For the full configuration reference, see [Configuration](../configuration/index.md).
