# OAuth2/OIDC Single Sign-On Configuration Guide

Open Website Defender (OWD) can act as an OIDC Provider, allowing any application that supports OAuth2/OIDC to authenticate users via the standard protocol for single sign-on (SSO).

## Prerequisites

- OWD is deployed and accessible via a public or internal network
- You have admin access to the OWD Admin dashboard

---

## Step 1: Configure OWD

Edit `config/config.yaml` and configure the `oauth` section:

```yaml
oauth:
  enabled: true
  # Required! Set to OWD's public-facing URL + ROOT_PATH
  # e.g., if OWD is at https://auth.example.com with default ROOT_PATH /wall
  issuer: "https://auth.example.com/wall"
  # Strongly recommended for production: specify a persistent RSA private key
  # If empty, a key is generated in memory on each restart (all tokens invalidated on restart)
  rsa-private-key-path: "/data/owd/rsa_private.pem"
  # Defaults below, adjust as needed
  authorization-code-lifetime: 300     # Auth code TTL in seconds (default: 5 minutes)
  access-token-lifetime: 3600          # Access token TTL (default: 1 hour)
  refresh-token-lifetime: 2592000      # Refresh token TTL (default: 30 days)
  id-token-lifetime: 3600              # ID token TTL (default: 1 hour)
```

### Generate RSA Private Key

```bash
openssl genrsa -out /data/owd/rsa_private.pem 2048
chmod 600 /data/owd/rsa_private.pem
```

!!! warning "Persist the RSA key in production"
    Without `rsa-private-key-path`, OWD generates a new key pair on every restart, invalidating all issued access tokens and ID tokens. Downstream applications will require re-authentication.

### Restart OWD

Restart OWD after changing the configuration.

### Verify OIDC Endpoints

```bash
# Check OIDC discovery document
curl https://auth.example.com/wall/.well-known/openid-configuration

# Check JWKS public key
curl https://auth.example.com/wall/.well-known/jwks.json
```

The discovery document should return:

```json
{
  "issuer": "https://auth.example.com/wall",
  "authorization_endpoint": "https://auth.example.com/wall/oauth/authorize",
  "token_endpoint": "https://auth.example.com/wall/oauth/token",
  "userinfo_endpoint": "https://auth.example.com/wall/oauth/userinfo",
  "jwks_uri": "https://auth.example.com/wall/.well-known/jwks.json",
  "response_types_supported": ["code"],
  "scopes_supported": ["openid", "profile", "email"],
  ...
}
```

---

## Step 2: Set User Emails

The OIDC userinfo endpoint returns the user's email. Downstream applications use it to match or create local users.

1. Log in to Admin dashboard → **USERS_DB**
2. Edit a user and fill in the **EMAIL** field (e.g., `admin@example.com`)
3. Save

---

## Step 3: Create an OAuth Client

1. Log in to Admin dashboard → **OAUTH_CLIENTS**
2. Click **[ NEW_CLIENT ]**
3. Fill in the form:

| Field | Description | Example |
|---|---|---|
| Name | Application display name | `My App` |
| Redirect URIs | OAuth callback URLs (one per line) | `https://app.example.com/oauth2/callback` |
| Scopes | Allowed scopes | `openid profile email` |
| Trusted | Skip user consent page (seamless SSO) | Recommended: checked |

4. Click Confirm
5. **Copy the Client ID and Client Secret immediately** — the secret is shown only once!

---

## Step 4: Configure Downstream Applications

For any application that supports OpenID Connect, configure the following settings:

| Setting | Value |
|---|---|
| Discovery URL | `https://auth.example.com/wall/.well-known/openid-configuration` |
| Authorization URL | `https://auth.example.com/wall/oauth/authorize` |
| Token URL | `https://auth.example.com/wall/oauth/token` |
| UserInfo URL | `https://auth.example.com/wall/oauth/userinfo` |
| JWKS URL | `https://auth.example.com/wall/.well-known/jwks.json` |
| Client ID | From Admin dashboard |
| Client Secret | From Admin dashboard |
| Scopes | `openid profile email` |

---

## Step 5: Nginx Configuration (Optional)

If the downstream application is behind OWD WAF protection (via nginx auth_request), ensure OAuth callbacks also pass through the WAF.

```nginx
# OWD itself
server {
    listen 443 ssl;
    server_name auth.example.com;

    location / {
        proxy_pass http://127.0.0.1:9999;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Downstream application (protected by OWD WAF)
server {
    listen 443 ssl;
    server_name app.example.com;

    # Every request goes through OWD /auth check (IP lists, WAF, rate limiting)
    auth_request /owd-auth;

    location = /owd-auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
        proxy_set_header X-Original-URI $request_uri;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Cookie $http_cookie;
    }

    # Redirect to OWD login on auth failure
    error_page 401 = @owd_login;
    location @owd_login {
        return 302 https://auth.example.com/wall/guard/login?redirect=$scheme://$host$request_uri;
    }

    location / {
        proxy_pass http://127.0.0.1:3000;  # Application internal address
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Login Flow

### Trusted Client (Seamless SSO)

```
User visits application → clicks "Sign in with OWD"
  → 302 to OWD /oauth/authorize
  → OWD checks cookie
    → Logged in: auto-issue auth code → 302 back to application callback
    → Not logged in: redirect to guard login → login → auto-issue auth code → 302 back
  → Application server exchanges code for tokens (POST /oauth/token)
  → Application fetches user info with access_token (GET /oauth/userinfo)
  → Application creates/matches local user, login complete
```

### Non-trusted Client

Same flow, but a consent page is shown before issuing the auth code, requiring user approval.

### Token Types

| Token | Format | Purpose |
|---|---|---|
| access_token | RS256-signed JWT | Call `/oauth/userinfo` to get user identity |
| id_token | RS256-signed JWT | Client-side identity parsing (no network request needed) |
| refresh_token | Random string (stored in DB) | Refresh expired access_token |

### UserInfo Response Example

```json
{
  "sub": "1",
  "preferred_username": "admin",
  "email": "admin@example.com",
  "email_verified": true
}
```

---

## Troubleshooting

### OIDC Discovery returns 404

- Verify `oauth.enabled: true` in config
- Verify the URL includes ROOT_PATH (default `/wall`)

### Token exchange fails (invalid_client)

- Verify Client ID and Client Secret are correct
- Verify the OAuth Client status is Active

### UserInfo returns 401

- Ensure you're using the access_token (not the id_token or refresh_token)
- Check that the token hasn't expired
- If OWD was restarted without a persistent RSA key, old tokens are invalid

### Redirect URI mismatch (invalid redirect_uri)

- Ensure the callback URL configured in the downstream application **exactly matches** what's registered in OWD (including protocol, path, trailing slash)

### User login creates a new account instead of matching existing

- Ensure the OWD user's Email matches the existing user's Email in the downstream app
- Some apps match by `preferred_username` — ensure the OWD username matches the target app username
