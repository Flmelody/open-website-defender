# 地域封锁

## 概述

Website Defender 支持基于国家/地区的地域封锁功能，使用 **MaxMind GeoLite2-Country** 数据库将请求 IP 映射到国家/地区代码，并根据封锁列表决定是否拒绝访问。

!!! info "GeoLite2 数据库"
    使用地域封锁功能需要下载 MaxMind GeoLite2-Country 数据库文件（`.mmdb` 格式）。您可以从 [MaxMind 官网](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) 免费注册并下载。

## 功能特点

- 基于 IP 地理位置的访问控制
- 使用标准 ISO 3166-1 国家/地区代码（如 `CN`、`US`、`RU`）
- 在中间件链中位于 WAF 之前执行，提前拦截来自封锁地区的请求
- 封锁的国家代码通过管理后台进行管理

## 管理封锁规则

### 通过管理后台

1. 登录管理后台
2. 进入**地域封锁**管理页面
3. 添加或删除要封锁的国家/地区代码

### 通过 API

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/geo-block-rules` | 添加封锁规则 |
| `GET` | `/geo-block-rules` | 查询封锁规则列表 |
| `DELETE` | `/geo-block-rules/:id` | 删除封锁规则 |

## 配置

在 `config/config.yaml` 中启用地域封锁并配置数据库路径：

```yaml
geo-blocking:
  enabled: false
  # MaxMind GeoLite2-Country.mmdb 文件路径
  database-path: "/path/to/GeoLite2-Country.mmdb"
```

!!! warning "注意事项"
    - 地域封锁功能默认关闭，需要手动启用
    - 必须提供有效的 GeoLite2-Country.mmdb 文件路径
    - MaxMind 定期更新数据库，建议定期下载最新版本以保持准确性

---

## 相关页面

- [配置说明](../configuration/index.md) - 完整的运行时配置参考
- [架构说明](../architecture/index.md) - 地域封锁在中间件链中的位置
- [API 参考](../api-reference/index.md) - 地域封锁管理 API
