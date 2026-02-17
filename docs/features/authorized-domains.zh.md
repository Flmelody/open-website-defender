# 授权域管理

授权域是 Website Defender 中所有受保护域名的集中注册表，作为 IP 白名单域名绑定和用户访问控制的**统一数据源**。

## 功能定位

在没有授权域管理的情况下，配置 IP 白名单条目或用户访问权限时需要手动输入域名，容易出错且不一致。授权域通过以下方式解决这些问题：

- 提供一个**集中管理**所有受保护域名的位置
- 为 IP 白名单和用户管理表单中的域名字段提供**下拉选择器**

- 实现**多租户访问控制**，允许不同用户通过同一个 Defender 实例访问不同的受保护服务

## 管理方式

### 通过管理后台

1. 进入**授权域管理**页面
2. 输入域名（如 `app.example.com`）添加新的授权域
3. 查看所有已注册的授权域及创建时间
4. 删除不再需要的授权域

### 通过 API

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/authorized-domains` | 查询授权域列表（分页） |
| `GET` | `/authorized-domains?all=true` | 查询全部授权域（用于下拉选项） |
| `POST` | `/authorized-domains` | 注册新的授权域 |
| `DELETE` | `/authorized-domains/:id` | 删除授权域 |

!!! note "唯一性约束"
    域名必须唯一。尝试注册已存在的域名将返回 `409 Conflict` 响应。

## 访问控制

授权域实现了**多租户访问控制**。每个用户可以被分配特定的授权域，限制其可以访问哪些受保护的服务。

### 工作原理

1. 当请求到达 `/auth` 时，Defender 从 `X-Forwarded-Host` 读取域名（回退到 `Host` 请求头）
2. 令牌/Git Token 认证成功后，检查用户的授权域是否匹配请求的域名
3. 如果域名不匹配任何授权域模式，返回 `403 Forbidden`

!!! info "适用场景"
    当您使用同一个 Defender 实例保护多个内部服务时，授权域访问控制可以限制每个用户只能访问被授权的服务。例如，开发人员只能访问 `app.example.com`，而运维人员可以同时访问 `app.example.com` 和 `app2.example.com`。

### 授权域模式

| 模式 | 匹配 | 不匹配 |
|------|------|--------|
| `app.example.com` | `app.example.com` | `other.example.com`、`sub.app.example.com` |
| `*.example.com` | `app.example.com`、`dev.example.com` | `example.com` |
| `app.example.com, *.internal.org` | `app.example.com`、`svc.internal.org` | `other.example.com` |
| *（空）* | 所有域名（不限制） | - |

### 规则说明

- **授权域为空** = 不限制访问（向后兼容已有用户）
- **管理员用户** 始终跳过授权域检查，无论其配置如何
- 匹配**不区分大小写**
- 匹配前会自动剥离端口号（`app.example.com:3000` 匹配 `app.example.com`）

!!! tip "通配符匹配"
    通配符 `*` 仅匹配一级子域名。例如 `*.example.com` 匹配 `app.example.com`，但不匹配 `example.com` 本身，也不匹配 `sub.app.example.com`。

!!! note "管理员用户"
    管理员用户始终拥有全部访问权限，不受授权域限制。如需限制某用户的访问范围，请确保该用户不具有管理员权限。

### Nginx 配置示例

要将域名信息传递给 Defender 用于访问控制，需配置 Nginx 通过 `X-Forwarded-Host` 转发 `Host` 请求头：

```nginx
server {
    server_name app.example.com;

    location / {
        auth_request /auth;

        # 将原始域名传递给 Defender 用于授权域检查
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
```

!!! warning "必须设置 X-Forwarded-Host"
    如果 Nginx 未配置 `proxy_set_header X-Forwarded-Host $host;`，Defender 将回退到 `Host` 请求头。在反向代理环境中，`Host` 请求头可能是内部地址而非用户请求的原始域名，导致授权域检查不正确。

完整的 Nginx 配置指南请参阅 [Nginx 配置](../deployment/nginx-setup.md)。

### 示例场景

假设组织有以下受保护服务：

| 服务 | 域名 |
|------|------|
| 应用 1 | `app1.internal.org` |
| 应用 2 | `app2.internal.org` |
| 应用 3 | `app3.internal.org` |

可以这样配置用户的授权域：

| 用户 | 授权域 | 可访问 |
|------|--------|--------|
| `alice` | `*.internal.org` | 所有服务 |
| `bob` | `app1.internal.org, app2.internal.org` | 仅应用 1 和应用 2 |
| `charlie` | `app3.internal.org` | 仅应用 3 |
| `admin` | *（任意）* | 所有服务（管理员跳过检查） |

## 与其他功能的集成

### IP 白名单

添加 IP 白名单条目时，**授权域**字段提供一个下拉选择器，选项来自授权域注册表。也可以手动输入自定义值。

详见 [IP 黑白名单](ip-lists.md)。

### 用户管理

配置用户访问权限时，**授权域**字段提供一个多选下拉选择器，选项来自授权域注册表。用户可以选择多个域名，也可以输入自定义模式（如 `*.example.com`）。

详见[用户管理](user-management.md)。

---

## 相关页面

- [IP 黑白名单](ip-lists.md) -- IP 白名单域名绑定
- [用户管理](user-management.md) -- 用户授权域配置
- [认证与访问控制](authentication.md) -- 完整的认证流程说明
- [Nginx 配置](../deployment/nginx-setup.md) -- 完整的 Nginx 配置指南
- [API 参考](../api-reference/index.md) -- 完整的 API 接口文档
