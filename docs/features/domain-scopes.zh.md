# 域名作用域

域名作用域实现了多租户访问控制，允许不同用户通过同一个 Defender 实例访问不同的受保护服务。

## 工作原理

1. 当请求到达 `/auth` 时，Defender 从 `X-Forwarded-Host` 读取域名（回退到 `Host` 请求头）
2. 令牌/Git Token 认证成功后，检查用户的作用域是否匹配请求的域名
3. 如果域名不匹配任何作用域模式，返回 `403 Forbidden`

!!! info "适用场景"
    当您使用同一个 Defender 实例保护多个内部服务时，域名作用域可以限制每个用户只能访问被授权的服务。例如，开发人员只能访问 `gitea.example.com`，而运维人员可以同时访问 `gitea.example.com` 和 `jenkins.example.com`。

## 授权域集成

域名作用域与[授权域管理](authorized-domains.md)协同工作。授权域注册表提供所有受保护域名的集中列表，在配置用户作用域时自动填充下拉选择器。删除授权域时，系统会自动从所有用户的作用域列表中移除该域名。

## 作用域模式

| 模式 | 匹配 | 不匹配 |
|------|------|--------|
| `gitea.com` | `gitea.com` | `gitlab.com`、`sub.gitea.com` |
| `*.example.com` | `app.example.com`、`dev.example.com` | `example.com` |
| `gitea.com, *.internal.org` | `gitea.com`、`app.internal.org` | `gitlab.com` |
| *（空）* | 所有域名（不限制） | - |

## 规则说明

- **作用域为空** = 不限制访问（向后兼容已有用户）
- **管理员用户** 始终跳过作用域检查，无论其作用域值如何
- 匹配**不区分大小写**
- 匹配前会自动剥离端口号（`gitea.com:3000` 匹配作用域 `gitea.com`）

!!! tip "通配符匹配"
    通配符 `*` 仅匹配一级子域名。例如 `*.example.com` 匹配 `app.example.com`，但不匹配 `example.com` 本身，也不匹配 `sub.app.example.com`。

!!! note "管理员用户"
    管理员用户始终拥有全部访问权限，不受域名作用域限制。如需限制某用户的访问范围，请确保该用户不具有管理员权限。

## Nginx 配置示例

要将域名信息传递给 Defender，需配置 Nginx 通过 `X-Forwarded-Host` 转发 `Host` 请求头：

```nginx
server {
    server_name gitea.example.com;

    location / {
        auth_request /auth;

        # 将原始域名传递给 Defender 用于作用域检查
        proxy_pass http://gitea-backend;
    }

    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
}
```

!!! warning "必须设置 X-Forwarded-Host"
    如果 Nginx 未配置 `proxy_set_header X-Forwarded-Host $host;`，Defender 将回退到 `Host` 请求头。在反向代理环境中，`Host` 请求头可能是内部地址而非用户请求的原始域名，导致作用域检查不正确。

---

## 相关页面

- [授权域管理](authorized-domains.md) - 集中管理受保护域名
- [认证与访问控制](authentication.md) - 完整的认证流程说明
- [用户管理](user-management.md) - 如何为用户设置域名作用域
- [Nginx 配置](../deployment/nginx-setup.md) - 完整的 Nginx 配置指南
