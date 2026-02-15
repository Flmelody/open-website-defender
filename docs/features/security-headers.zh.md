# 安全响应头

## 概述

Website Defender 自动为所有响应附加安全头，增强浏览器端的安全防护。这些安全头有助于防范常见的 Web 攻击，如点击劫持、MIME 类型混淆、XSS 等。

## 安全头列表

| 响应头 | 值 | 说明 |
|--------|------|------|
| `X-Content-Type-Options` | `nosniff` | 阻止浏览器进行 MIME 类型嗅探，防止将非脚本文件作为脚本执行 |
| `X-XSS-Protection` | `1; mode=block` | 启用浏览器内置的 XSS 过滤器，检测到 XSS 攻击时阻止页面渲染 |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | 控制 Referer 请求头的发送策略：同源请求发送完整 URL，跨域请求仅发送源（origin） |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | 禁止页面使用摄像头、麦克风和地理位置 API，降低隐私泄露风险 |
| `X-Frame-Options` | 可配置（默认 `DENY`） | 控制页面是否可被嵌入 `<iframe>`。`DENY` 禁止所有嵌入，`SAMEORIGIN` 允许同源嵌入 |
| `Strict-Transport-Security` | 可选 | HSTS 头，强制浏览器通过 HTTPS 访问。仅在启用 HSTS 配置时添加 |

## 配置

安全响应头的部分选项可通过 `config/config.yaml` 配置：

```yaml
security:
  headers:
    # 启用 HSTS（仅在确认使用 HTTPS 时开启）
    hsts-enabled: false
    # X-Frame-Options: DENY、SAMEORIGIN 或留空禁用
    frame-options: "DENY"
```

!!! warning "HSTS 注意事项"
    仅在您的站点完全通过 HTTPS 提供服务时才启用 HSTS。一旦启用，浏览器将在指定时间内强制使用 HTTPS 访问，错误配置可能导致用户无法访问站点。

!!! info "X-Frame-Options 选项说明"
    - `DENY`：完全禁止页面被嵌入 iframe（最安全）
    - `SAMEORIGIN`：仅允许同源页面嵌入
    - 留空：不发送此响应头（不推荐）

---

## 相关页面

- [配置说明](../configuration/index.md) - 安全头相关配置项
- [架构说明](../architecture/index.md) - 安全响应头在中间件链中的位置
