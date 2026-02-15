# WAF 规则

## 概述

Website Defender 内置了 Web 应用防火墙（WAF），提供基于正则表达式的请求过滤功能。WAF 会检查以下请求内容：

- **URL 路径**
- **查询参数**
- **User-Agent**
- **请求体**（最大 10KB）

每条规则支持两种动作：

| 动作 | 行为 |
|------|------|
| `block` | 返回 `403 Forbidden`，拦截请求 |
| `log` | 放行请求，但记录匹配日志 |

!!! tip "建议"
    对于新添加的自定义规则，建议先使用 `log` 动作观察一段时间，确认没有误报后再切换为 `block` 动作。

## 内置规则

Defender 提供 9 条开箱即用的内置规则，覆盖常见的 Web 攻击类型：

### SQL 注入

| 规则 | 说明 |
|------|------|
| Union Select | 检测 `UNION SELECT` 联合查询攻击 |
| Common Patterns | 检测 `; DROP`、`; ALTER`、`; DELETE` 等破坏性语句 |
| Boolean Injection | 检测 `' OR 1=1` 等布尔型盲注 |
| Comment Injection | 检测 `' --` 和 `/* */` 注释注入 |

### XSS（跨站脚本攻击）

| 规则 | 说明 |
|------|------|
| Script Tag | 检测 `<script>` 标签注入 |
| Event Handler | 检测 `onerror=`、`onclick=` 等事件属性 |
| JavaScript Protocol | 检测 `javascript:` 和 `vbscript:` 协议 |

### 路径穿越

| 规则 | 说明 |
|------|------|
| Dot Dot Slash | 检测 `../`、`..\` 及 URL 编码变体 |
| Sensitive Files | 检测访问 `/etc/passwd`、`/proc/self` 等敏感文件 |

## 自定义规则

除了内置规则，您还可以通过管理后台添加自定义 WAF 规则：

1. 登录管理后台
2. 进入 **WAF 规则** 管理页面
3. 点击添加新规则
4. 配置规则名称、正则表达式模式和动作（`block` 或 `log`）

也可以通过 API 管理 WAF 规则，详见 [API 参考](../api-reference/index.md)。

!!! warning "正则表达式性能"
    请谨慎编写正则表达式，避免使用可能导致回溯爆炸的复杂模式。性能不佳的正则表达式可能影响请求处理速度。

## 配置

WAF 功能通过 `config/config.yaml` 中的 `request-filtering` 配置项控制：

```yaml
# WAF（SQL 注入、XSS、路径穿越检测）
request-filtering:
  enabled: true
```

!!! note "启用与禁用"
    将 `enabled` 设置为 `false` 可完全关闭 WAF 功能。关闭后，所有 WAF 规则（包括内置规则和自定义规则）均不会执行。

---

## 相关页面

- [访问日志](access-logs.md) - 查看 WAF 拦截记录
- [配置说明](../configuration/index.md) - WAF 配置项详解
- [API 参考](../api-reference/index.md) - WAF 规则管理 API
