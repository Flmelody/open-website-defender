# OAuth2/OIDC 单点登录配置指南

Open Website Defender (OWD) 可作为 OIDC Provider，让任何支持 OAuth2/OIDC 的应用对接 OWD，实现单点登录 (SSO)。

## 前置条件

- OWD 已部署并可通过公网/内网访问
- 已有管理员账号登录 Admin 后台

---

## 第一步：配置 OWD

编辑 `config/config.yaml`，确认 `oauth` 段配置正确：

```yaml
oauth:
  enabled: true
  # 必须设置！填写 OWD 的公网访问地址 + ROOT_PATH
  # 例如 OWD 部署在 https://auth.example.com，ROOT_PATH 默认 /wall
  issuer: "https://auth.example.com/wall"
  # 生产环境强烈建议指定固定的 RSA 私钥文件
  # 留空则每次重启自动生成（重启后所有已签发 token 失效）
  rsa-private-key-path: "/data/owd/rsa_private.pem"
  # 以下为默认值，按需调整
  authorization-code-lifetime: 300     # 授权码有效期（秒），默认 5 分钟
  access-token-lifetime: 3600          # access token 有效期，默认 1 小时
  refresh-token-lifetime: 2592000      # refresh token 有效期，默认 30 天
  id-token-lifetime: 3600              # id token 有效期，默认 1 小时
```

### 生成 RSA 私钥

```bash
openssl genrsa -out /data/owd/rsa_private.pem 2048
chmod 600 /data/owd/rsa_private.pem
```

!!! warning "生产环境必须持久化 RSA 私钥"
    如果不指定 `rsa-private-key-path`，OWD 每次重启会生成新的密钥对，导致所有已签发的 access token 和 id token 失效，下游应用需要重新认证。

### 重启 OWD

修改配置后重启 OWD 使配置生效。

### 验证 OIDC 端点

```bash
# 检查 OIDC 发现文档
curl https://auth.example.com/wall/.well-known/openid-configuration

# 检查 JWKS 公钥
curl https://auth.example.com/wall/.well-known/jwks.json
```

发现文档应返回类似：

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

## 第二步：给用户设置 Email

OIDC userinfo 接口会返回用户的 email，下游应用用它来匹配或创建本地用户。

1. 登录 Admin 后台 → **USERS_DB**
2. 编辑用户，填写 **EMAIL** 字段（如 `admin@example.com`）
3. 保存

---

## 第三步：创建 OAuth Client

1. 登录 Admin 后台 → **OAUTH_CLIENTS**
2. 点击 **[ NEW_CLIENT ]**
3. 填写表单：

| 字段 | 说明 | 示例 |
|---|---|---|
| 名称 | 应用显示名称 | `My App` |
| 回调地址 | 应用的 OAuth 回调 URL，每行一个 | `https://app.example.com/oauth2/callback` |
| 权限范围 | 允许的 scope | `openid profile email` |
| 信任客户端 | 勾选后跳过用户授权确认页（无感登录） | 建议勾选 |

4. 点击确认
5. **立即复制 Client ID 和 Client Secret** — Secret 仅显示一次！

---

## 第四步：配置下游应用

对于任何支持 OpenID Connect 的应用，配置以下信息：

| 配置项 | 值 |
|---|---|
| Discovery URL | `https://auth.example.com/wall/.well-known/openid-configuration` |
| Authorization URL | `https://auth.example.com/wall/oauth/authorize` |
| Token URL | `https://auth.example.com/wall/oauth/token` |
| UserInfo URL | `https://auth.example.com/wall/oauth/userinfo` |
| JWKS URL | `https://auth.example.com/wall/.well-known/jwks.json` |
| Client ID | Admin 后台创建的 Client ID |
| Client Secret | Admin 后台创建的 Client Secret |
| Scopes | `openid profile email` |

---

## 第五步：Nginx 配置（可选）

如果下游应用在 OWD WAF 保护下（通过 nginx auth_request），需要确保 OAuth 回调也能通过 WAF。

```nginx
# OWD 自身
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

# 下游应用（受 OWD WAF 保护）
server {
    listen 443 ssl;
    server_name app.example.com;

    # 每个请求先经过 OWD /auth 检查（IP黑白名单、WAF、限流）
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

    # auth 失败时跳转 OWD 登录页
    error_page 401 = @owd_login;
    location @owd_login {
        return 302 https://auth.example.com/wall/guard/login?redirect=$scheme://$host$request_uri;
    }

    location / {
        proxy_pass http://127.0.0.1:3000;  # 应用内部地址
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 登录流程说明

### Trusted Client（无感登录）

```
用户访问应用 → 点击 "Sign in with OWD"
  → 302 到 OWD /oauth/authorize
  → OWD 检查 cookie
    → 已登录：自动签发 auth code → 302 回应用 callback
    → 未登录：跳转 guard 登录页 → 登录后 → 自动签发 auth code → 302 回应用 callback
  → 应用服务端用 code 换 token（POST /oauth/token）
  → 应用用 access_token 获取用户信息（GET /oauth/userinfo）
  → 应用创建/匹配本地用户，登录完成
```

### Non-trusted Client

流程相同，但在签发 auth code 前会显示授权确认页，要求用户手动批准。

### Token 说明

| Token | 格式 | 用途 |
|---|---|---|
| access_token | RS256 签名的 JWT | 调用 `/oauth/userinfo` 获取用户身份 |
| id_token | RS256 签名的 JWT | 客户端直接解析用户身份（无需网络请求） |
| refresh_token | 随机字符串（存 DB） | 过期后刷新 access_token |

### UserInfo 返回示例

```json
{
  "sub": "1",
  "preferred_username": "admin",
  "email": "admin@example.com",
  "email_verified": true
}
```

---

## 故障排查

### OIDC Discovery 返回 404

- 确认 `oauth.enabled: true`
- 确认访问地址包含 ROOT_PATH（默认 `/wall`）

### Token 交换失败 (invalid_client)

- 确认 Client ID 和 Client Secret 正确
- 确认 OAuth Client 状态为 Active

### UserInfo 返回 401

- 确认使用的是 access_token（不是 id_token 或 refresh_token）
- 确认 token 未过期
- 如果重启过 OWD 且未配置持久化 RSA 私钥，旧 token 已失效

### 回调地址不匹配 (invalid redirect_uri)

- 确认下游应用配置的回调地址与 OWD 中注册的**完全一致**（包括协议、路径、末尾斜杠）

### 用户登录后创建了新用户而非匹配已有用户

- 确认 OWD 用户的 Email 与下游应用已有用户的 Email 一致
- 部分应用使用 `preferred_username` 匹配，确认 OWD 用户名与目标应用用户名一致
