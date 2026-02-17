# Nginx 配置

## 概述

Website Defender 设计为配合 Nginx 的 `auth_request` 模块使用。本页面提供完整的 Nginx 配置示例和详细说明。

## 完整配置示例

```nginx
server {
    listen 80;
    server_name app.example.com;

    location / {
        # 在转发请求之前，先向 Defender 发起认证子请求
        auth_request /auth;

        # 认证失败时的错误处理
        # 401: 重定向到登录页面
        error_page 401 = @login;
        # 403: 显示拒绝访问页面
        error_page 403 = @forbidden;

        # 认证通过后，将请求代理到实际的内部应用
        proxy_pass http://app-backend;
    }

    # Defender 认证端点（内部子请求）
    location = /auth {
        # 标记为内部请求，外部无法直接访问
        internal;

        # 转发认证请求到 Defender
        proxy_pass http://127.0.0.1:9999/wall/auth;

        # 传递原始域名（用于授权域检查）
        proxy_set_header X-Forwarded-Host $host;

        # 传递客户端真实 IP（用于 IP 黑白名单和访问日志）
        proxy_set_header X-Forwarded-For $remote_addr;

        # 传递原始请求 URI（用于识别 Git 请求）
        proxy_set_header X-Original-URI $request_uri;

        # auth_request 不需要请求体
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }

    # 登录页面重定向
    location @login {
        return 302 http://127.0.0.1:9999/wall/guard/;
    }

    # 拒绝访问页面
    location @forbidden {
        return 403;
    }
}
```

## 指令详解

### auth_request

```nginx
auth_request /auth;
```

Nginx 的 `auth_request` 指令在处理原始请求之前，先向指定的内部 URI 发起子请求。如果子请求返回 `2xx` 状态码，原始请求继续处理；如果返回 `401` 或 `403`，原始请求将被拒绝。

### internal

```nginx
internal;
```

`internal` 指令确保该 location 只能通过 Nginx 内部子请求访问，外部客户端无法直接请求 `/auth` 路径。

### X-Forwarded-Host

```nginx
proxy_set_header X-Forwarded-Host $host;
```

将客户端请求的原始域名（`Host` 请求头）传递给 Defender。Defender 使用此信息进行[授权域](../features/authorized-domains.md)检查。

!!! warning "必须配置"
    如果使用了授权域访问控制功能，此配置项是必需的。没有此配置，Defender 将无法正确判断请求的目标域名。

### X-Forwarded-For

```nginx
proxy_set_header X-Forwarded-For $remote_addr;
```

将客户端的真实 IP 地址传递给 Defender。Defender 使用此信息进行：

- IP 黑名单/白名单检查
- 速率限制
- 访问日志记录
- 地域封锁

!!! warning "真实 IP 重要性"
    如果不配置此项，Defender 看到的将是 Nginx 的 IP 地址（通常是 `127.0.0.1`），导致 IP 相关的安全功能全部失效。

### X-Original-URI

```nginx
proxy_set_header X-Original-URI $request_uri;
```

将客户端的原始请求 URI 传递给 Defender。Defender 使用此信息识别 Git HTTP 请求，仅对 Git 请求进行 Git Token 认证。

### proxy_pass_request_body off

```nginx
proxy_pass_request_body off;
proxy_set_header Content-Length "";
```

认证子请求不需要原始请求体。关闭请求体转发可以减少不必要的数据传输，提高认证效率。

## 多域名配置

如果您使用同一个 Defender 实例保护多个内部服务，可以为每个域名配置独立的 server 块：

```nginx
# 服务 1
server {
    server_name app.example.com;

    location / {
        auth_request /auth;
        proxy_pass http://app-backend;
    }

    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Original-URI $request_uri;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
}

# 服务 2
server {
    server_name app2.example.com;

    location / {
        auth_request /auth;
        proxy_pass http://app2-backend;
    }

    location = /auth {
        internal;
        proxy_pass http://127.0.0.1:9999/wall/auth;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Original-URI $request_uri;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
    }
}
```

!!! tip "授权域访问控制"
    配合[授权域](../features/authorized-domains.md)访问控制功能，您可以精确控制每个用户可以访问哪些受保护的服务。

---

## 相关页面

- [架构说明](../architecture/index.md) - 了解 Defender 与 Nginx 的集成架构
- [授权域管理](../features/authorized-domains.md) - 多租户访问控制
- [部署指南](index.md) - 部署概述和检查清单
