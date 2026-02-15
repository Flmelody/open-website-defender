# API 参考

## 概述

Website Defender 提供 RESTful API 用于管理所有功能。所有路由均以可配置的 `ROOT_PATH`（默认 `/wall`）为前缀。

!!! info "认证说明"
    标记为"需要鉴权"的 API 需要在请求中携带有效的 JWT 令牌（通过 `Defender-Authorization` 请求头或 `flmelody.token` Cookie）。

## 完整 API 参考

### 公开端点

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/login` | 用户登录，返回 JWT 令牌 | 否 |
| `GET` | `/auth` | 验证凭证（IP 名单 + 令牌），供 Nginx `auth_request` 调用 | 否 |
| `GET` | `/health` | 健康检查 | 否 |

### 仪表盘

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `GET` | `/dashboard/stats` | 获取仪表盘统计数据（请求数、拦截数、运行时间等） | 是 |

### 用户管理

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/users` | 创建用户 | 是 |
| `GET` | `/users` | 查询用户列表 | 是 |
| `PUT` | `/users/:id` | 更新用户信息 | 是 |
| `DELETE` | `/users/:id` | 删除用户 | 是 |

### IP 黑名单

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/ip-black-list` | 添加黑名单条目 | 是 |
| `GET` | `/ip-black-list` | 查询黑名单列表 | 是 |
| `DELETE` | `/ip-black-list/:id` | 删除黑名单条目 | 是 |

### IP 白名单

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/ip-white-list` | 添加白名单条目 | 是 |
| `GET` | `/ip-white-list` | 查询白名单列表 | 是 |
| `DELETE` | `/ip-white-list/:id` | 删除白名单条目 | 是 |

### WAF 规则

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/waf-rules` | 添加 WAF 规则 | 是 |
| `GET` | `/waf-rules` | 查询 WAF 规则列表 | 是 |
| `PUT` | `/waf-rules/:id` | 更新 WAF 规则 | 是 |
| `DELETE` | `/waf-rules/:id` | 删除 WAF 规则 | 是 |

### 访问日志

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `GET` | `/access-logs` | 查询访问日志列表 | 是 |
| `GET` | `/access-logs/stats` | 获取访问日志统计数据 | 是 |

### 地域封锁

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/geo-block-rules` | 添加地域封锁规则 | 是 |
| `GET` | `/geo-block-rules` | 查询地域封锁规则列表 | 是 |
| `DELETE` | `/geo-block-rules/:id` | 删除地域封锁规则 | 是 |

### 许可证管理

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/licenses` | 创建许可证 | 是 |
| `GET` | `/licenses` | 查询许可证列表 | 是 |
| `DELETE` | `/licenses/:id` | 删除许可证 | 是 |

### 系统设置

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `GET` | `/system/settings` | 获取系统设置 | 是 |
| `PUT` | `/system/settings` | 更新系统设置 | 是 |
| `POST` | `/system/reload` | 重载配置并清除缓存 | 是 |

!!! tip "路径前缀"
    以上所有路径均需加上 `ROOT_PATH` 前缀。例如，如果 `ROOT_PATH` 为默认值 `/wall`，则登录接口的完整路径为 `/wall/login`。

---

## 相关页面

- [认证与访问控制](../features/authentication.md) - 认证方式详解
- [配置说明](../configuration/index.md) - ROOT_PATH 等路径配置
- [Nginx 配置](../deployment/nginx-setup.md) - 配置 Nginx 代理 API 请求
