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

## 语义分析引擎

除了基于正则表达式的模式匹配，Website Defender 还内置了**语义分析引擎**，能够理解 SQL 和 HTML 的结构，而非仅依赖字符串模式匹配。该引擎提供了更深层的检测能力，可以捕获绕过正则规则的攻击，同时显著降低误报率。

### 工作原理

语义分析引擎采用受 libinjection 启发的多阶段流水线：

1. **词法分析（Tokenization）** -- 使用专用 SQL 词法分析器将输入分解为类型化的 Token 流（关键字、字符串、数字、运算符、注释、函数等）。
2. **Token 折叠（Folding）** -- 合并复合关键字（如 `UNION ALL`、`GROUP BY`），折叠算术表达式，吸收一元运算符。Token 流被压缩至最多 5 个 Token。
3. **指纹生成（Fingerprint）** -- 每个 Token 贡献一个类型字符，生成紧凑的指纹字符串（例如 `' OR 1=1` 生成 `s&1o1`）。
4. **指纹匹配** -- 将指纹与已知 SQL 注入攻击模式的精选集合进行比对，覆盖 UNION 注入、布尔注入、堆叠查询、注释注入、函数攻击等。
5. **白名单过滤** -- 看起来像自然语言的匹配结果（全部为普通单词 Token，无 SQL 特有运算符）会被过滤，以减少误报。

对于 XSS 检测，引擎执行 HTML 上下文分析而非简单的正则匹配。它检测 script 标签、标签上下文中的事件处理属性、危险协议（`javascript:`、`vbscript:`）以及可执行的 HTML 标签（带有事件处理器的 `<iframe>`、`<object>`、`<svg>` 等）。

### 双重检测模式

语义引擎同时以两种模式运行：

- **正则确认模式** -- 当 `sqli` 或 `xss` 类别的正则 WAF 规则匹配时，语义引擎会进行二次确认。如果语义分析未确认该匹配，请求将被视为误报并放行。
- **独立检测模式** -- 在所有正则规则评估完毕后，语义引擎独立扫描所有请求字段（路径、查询参数、User-Agent、请求体、请求头、Cookie），检测正则规则可能完全遗漏的 SQLi 和 XSS 攻击。

### 配置

可以在 `config/config.yaml` 中启用语义分析：

```yaml
request-filtering:
  enabled: true
  semantic-analysis:
    enabled: true
```

也可以在管理后台的**系统设置**中实时开关，或通过系统设置 API 进行配置（`PUT /system/settings`，设置 `semantic_analysis_enabled: true`）。

!!! tip "生产环境建议"
    建议在生产环境中启用语义分析。它能减少正则规则产生的误报（通过结构化确认），并能捕获使用编码技巧或语法变体来绕过正则模式的复杂攻击。

## WAF 排除规则

WAF 排除规则允许特定请求路径跳过 WAF 检查，可以全局生效或仅针对特定规则。这对于防止已知安全端点的误报非常有用，例如合法接受类 SQL 输入的 API 路由、富文本编辑器或 Webhook 接收器。

### 工作原理

当 WAF 规则匹配到请求时，系统会在执行任何操作之前检查排除列表。如果请求路径匹配到排除规则，该 WAF 规则将被跳过。排除规则支持三种匹配运算符：

| 运算符 | 行为 |
|--------|------|
| **prefix** | 请求路径以排除路径开头时匹配（默认） |
| **exact** | 请求路径与排除路径完全相同时匹配 |
| **regex** | 请求路径匹配排除规则的正则表达式时匹配 |

### 作用范围

每条排除规则可以按以下两种方式确定作用范围：

- **全局**（rule_id = 0） -- 排除规则适用于所有 WAF 规则。匹配路径的请求将跳过整个 WAF 检查。
- **指定规则** -- 排除规则绑定到特定 WAF 规则 ID，仅跳过该规则的检查，其他规则继续生效。

### 管理排除规则

排除规则可以通过管理后台的 WAF 管理页面进行配置，也可以通过 API 管理：

| 方法 | 端点 | 说明 |
|------|------|------|
| `POST` | `/waf-exclusions` | 创建新的排除规则 |
| `GET` | `/waf-exclusions` | 查询排除规则列表（分页） |
| `DELETE` | `/waf-exclusions/:id` | 删除排除规则 |

示例 -- 为 API 端点创建前缀排除规则：

```json
{
  "path": "/api/webhooks/",
  "operator": "prefix",
  "rule_id": 0,
  "enabled": true
}
```

!!! warning "谨慎使用排除规则"
    每条排除规则都会在 WAF 防护中创建一个缺口。建议优先使用指定规则的排除而非全局排除，并使用尽可能精确的路径匹配。请定期审查排除规则，确保它们仍然必要。

## 配置

WAF 功能通过 `config/config.yaml` 中的 `request-filtering` 配置项控制：

```yaml
# WAF（SQL 注入、XSS、路径穿越检测）
request-filtering:
  enabled: true
  semantic-analysis:
    enabled: true
```

!!! note "启用与禁用"
    将 `request-filtering.enabled` 设置为 `false` 可完全关闭 WAF 功能。关闭后，所有 WAF 规则（包括内置规则和自定义规则）均不会执行。`semantic-analysis.enabled` 设置项独立控制语义分析引擎的开关。

---

## 相关页面

- [访问日志](access-logs.md) - 查看 WAF 拦截记录
- [配置说明](../configuration/index.md) - WAF 配置项详解
- [API 参考](../api-reference/index.md) - WAF 规则管理 API
