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

## Semantic Analysis Engine

Beyond regex-based pattern matching, Website Defender includes a **semantic analysis engine** that understands the structure of SQL and HTML rather than relying solely on string patterns. This provides a deeper detection layer that catches attacks designed to bypass regex rules while significantly reducing false positives.

### How It Works

The semantic analysis engine uses a multi-stage pipeline inspired by libinjection:

1. **Tokenization** -- The input is broken into a stream of typed tokens (keywords, strings, numbers, operators, comments, functions, etc.) using a purpose-built SQL lexer.
2. **Token Folding** -- Compound keywords are merged (e.g., `UNION ALL`, `GROUP BY`), arithmetic expressions are collapsed, and unary operators are absorbed. The token stream is reduced to at most 5 tokens.
3. **Fingerprint Generation** -- Each token contributes a single type character, producing a compact fingerprint string (e.g., `s&1o1` for `' OR 1=1`).
4. **Fingerprint Matching** -- The fingerprint is checked against a curated set of known SQL injection attack patterns covering UNION-based injection, boolean injection, stacked queries, comment injection, function-based attacks, and more.
5. **Whitelist Filtering** -- Matches that appear to be natural language (all bareword tokens with no SQL-specific operators) are suppressed to reduce false positives.

For XSS detection, the engine performs HTML context analysis rather than simple regex matching. It checks for script tags, event handler attributes within tag contexts, dangerous protocol handlers (`javascript:`, `vbscript:`), and executable HTML tags (`<iframe>`, `<object>`, `<svg>` with event handlers, etc.).

### Dual-Mode Operation

The semantic engine operates in two modes simultaneously:

- **Regex Confirmation Mode** -- When a regex WAF rule in the `sqli` or `xss` category matches, the semantic engine is consulted to confirm the match. If semantic analysis does not confirm the match, the request is treated as a false positive and allowed through.
- **Independent Detection Mode** -- After all regex rules have been evaluated, the semantic engine independently scans all request fields (path, query, user-agent, body, headers, cookies) for SQLi and XSS patterns that the regex rules may have missed entirely.

### Configuration

Semantic analysis can be enabled in `config/config.yaml`:

```yaml
request-filtering:
  enabled: true
  semantic-analysis:
    enabled: true
```

It can also be toggled at runtime through the admin dashboard under **System Settings**, or via the system settings API (`PUT /system/settings` with `semantic_analysis_enabled: true`).

!!! tip "Recommended for Production"
    Enabling semantic analysis is recommended for production deployments. It reduces false positives from regex rules (by requiring structural confirmation) and catches sophisticated attacks that use encoding tricks or syntax variations to evade regex patterns.

## WAF Exclusions

WAF exclusions allow specific request paths to bypass WAF checks, either globally or for individual rules. This is useful for preventing false positives on known-safe endpoints such as API routes that legitimately accept SQL-like input, rich text editors, or webhook receivers.

### How Exclusions Work

When a WAF rule matches a request, the exclusion list is checked before any action is taken. If the request path matches an exclusion, the rule is skipped for that request. Exclusions are evaluated using one of three matching operators:

| Operator | Behavior |
|----------|----------|
| **prefix** | Matches if the request path starts with the exclusion path (default) |
| **exact** | Matches only if the request path is exactly equal to the exclusion path |
| **regex** | Matches if the request path matches the exclusion's regular expression |

### Scope

Each exclusion can be scoped in one of two ways:

- **Global** (rule ID = 0) -- The exclusion applies to all WAF rules. Any request matching the path will bypass the entire WAF.
- **Rule-specific** -- The exclusion is tied to a specific WAF rule by its ID. Only that rule is bypassed; other rules continue to apply.

### Managing Exclusions

Exclusions can be managed through the admin dashboard under the WAF section, or via the API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/waf-exclusions` | Create a new exclusion |
| `GET` | `/waf-exclusions` | List all exclusions (paginated) |
| `DELETE` | `/waf-exclusions/:id` | Delete an exclusion |

Example -- creating a prefix exclusion for an API endpoint:

```json
{
  "path": "/api/webhooks/",
  "operator": "prefix",
  "rule_id": 0,
  "enabled": true
}
```

!!! warning "Use Exclusions Sparingly"
    Each exclusion creates a gap in WAF coverage. Prefer rule-specific exclusions over global ones, and use the narrowest path match possible. Review exclusions periodically to ensure they are still needed.

## Configuration

The WAF can be enabled or disabled globally in `config/config.yaml`:

```yaml
request-filtering:
  enabled: true
  semantic-analysis:
    enabled: true
```

When `request-filtering.enabled` is set to `false`, no request inspection is performed and all WAF rules are skipped. The `semantic-analysis.enabled` setting controls the semantic analysis engine independently.

For the full configuration reference, see [Configuration](../configuration/index.md).
