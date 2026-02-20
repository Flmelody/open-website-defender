# JS Challenge (Proof-of-Work)

Website Defender can serve a JavaScript-based Proof-of-Work challenge to visitors, effectively filtering out automated bots and simple scripts that cannot execute JavaScript.

## How It Works

1. A visitor's browser receives an HTML page with an embedded JavaScript challenge
2. The browser computes a SHA256 Proof-of-Work (finding a nonce that produces a hash with a required number of leading zeros)
3. Upon successful computation, a signed cookie (`_defender_pow`) is set
4. The cookie is valid for 24 hours (configurable) and bound to the visitor's IP
5. Subsequent requests with a valid cookie bypass the challenge

## Challenge Modes

| Mode | Behavior |
|------|----------|
| `off` | JS Challenge is disabled |
| `suspicious` | Only challenges IPs with a [threat score](threat-detection.md) >= 10 |
| `all` | Challenges all new visitors without a valid pass cookie |

!!! tip "Recommended Mode"
    The `suspicious` mode is recommended for most deployments. It only challenges visitors who have already exhibited suspicious behavior, minimizing impact on legitimate users.

## Bypassed Clients

The following requests skip the JS challenge automatically:

- **Whitelisted IPs** -- IPs on the whitelist are always exempt
- **Authenticated requests** -- Requests with a valid `Defender-Authorization` header
- **Git/License tokens** -- Requests with configured git or license token headers
- **Non-browser clients** -- Clients identified as `git`, `curl`, `wget`, etc.
- **Auth subrequests** -- The `/auth` endpoint used by Nginx `auth_request`

## Difficulty

The difficulty setting controls the computational effort required to solve the challenge:

| Difficulty | Leading Zeros | Approximate Iterations |
|------------|--------------|----------------------|
| 1 | 1 | ~16 |
| 2 | 2 | ~256 |
| 3 | 3 | ~4,096 |
| **4** (default) | **4** | **~65,536** |
| 5 | 5 | ~1,048,576 |
| 6 | 6 | ~16,777,216 |

Higher difficulty means more computation time for the client. The default value of 4 provides a good balance between bot deterrence and user experience (typically completes in under 2 seconds on modern devices).

## Configuration

```yaml
js-challenge:
  enabled: false
  # Mode: off | suspicious | all
  mode: "suspicious"
  # Difficulty: number of leading zeros in SHA256 hash (1-6)
  difficulty: 4
  # Pass cookie TTL in seconds (default: 24 hours)
  cookie-ttl: 86400
  # HMAC secret for signing cookies (auto-generated if empty)
  cookie-secret: ""
```

JS Challenge settings can also be configured via the admin dashboard under **System Settings**.

!!! warning "Cookie Secret in Production"
    If `cookie-secret` is left empty, a random secret is generated on each restart, invalidating all existing pass cookies. Set a stable secret in production.

---

## Related Pages

- [Threat Detection](threat-detection.md) -- How threat scores drive the `suspicious` mode
- [Security Events](security-events.md) -- JS challenge failures are recorded as security events
- [Configuration](../configuration/index.md) -- Full configuration reference
