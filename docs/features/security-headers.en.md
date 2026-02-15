# Security Headers

Website Defender automatically adds security-hardening HTTP headers to all responses. These headers instruct browsers to enable built-in security mechanisms and restrict potentially dangerous behaviors.

## Headers

| Header | Value | Description |
|--------|-------|-------------|
| `X-Content-Type-Options` | `nosniff` | Prevents browsers from MIME-type sniffing, forcing them to respect the declared `Content-Type` |
| `X-XSS-Protection` | `1; mode=block` | Enables the browser's built-in XSS filter and blocks the page if an attack is detected |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Sends the full URL as referrer for same-origin requests but only the origin for cross-origin requests |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | Disables access to camera, microphone, and geolocation APIs for the page and all embedded iframes |
| `X-Frame-Options` | Configurable (default: `DENY`) | Controls whether the page can be embedded in frames. `DENY` prevents all framing; `SAMEORIGIN` allows same-origin framing |
| `Strict-Transport-Security` | Optional (HSTS) | When enabled, instructs browsers to only access the site over HTTPS for the specified duration |

## Configuration

The `X-Frame-Options` and HSTS headers are configurable in `config/config.yaml`:

```yaml
security:
  headers:
    # Enable HSTS (only enable if your site is served over HTTPS)
    hsts-enabled: false
    # X-Frame-Options: DENY, SAMEORIGIN, or empty to disable
    frame-options: "DENY"
```

!!! warning "HSTS Caution"
    Only enable HSTS (`hsts-enabled: true`) if your site is **always** served over HTTPS. Once a browser receives an HSTS header, it will refuse to connect over plain HTTP for the duration of the policy. Enabling HSTS without HTTPS will make your site inaccessible.

!!! note "Other Headers"
    The `X-Content-Type-Options`, `X-XSS-Protection`, `Referrer-Policy`, and `Permissions-Policy` headers are always applied and are not configurable. They represent security best practices with no known downsides.

## Middleware Position

Security headers are applied as the **first middleware** in the [middleware chain](../architecture/index.md):

```
SecurityHeaders → CORS → BodyLimit → AccessLog → GeoBlock → WAF → RateLimiter → Route Handler
```

This ensures that every response -- including error responses -- includes the appropriate security headers.
