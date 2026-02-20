# JS 挑战（工作量证明）

Website Defender 可以向访问者提供基于 JavaScript 的工作量证明（Proof-of-Work）挑战，有效过滤无法执行 JavaScript 的自动化机器人和简单脚本。

## 工作原理

1. 访问者的浏览器接收到一个包含 JavaScript 挑战的 HTML 页面
2. 浏览器计算 SHA256 工作量证明（找到一个使哈希值具有指定数量前导零的随机数）
3. 计算成功后，设置一个签名 Cookie（`_defender_pow`）
4. Cookie 有效期为 24 小时（可配置），并绑定到访问者的 IP
5. 后续携带有效 Cookie 的请求将跳过挑战

## 挑战模式

| 模式 | 行为 |
|------|------|
| `off` | 禁用 JS 挑战 |
| `suspicious` | 仅对[威胁评分](threat-detection.md) >= 10 的 IP 发起挑战 |
| `all` | 对所有没有有效通行 Cookie 的新访问者发起挑战 |

!!! tip "推荐模式"
    大多数部署场景推荐使用 `suspicious` 模式。它仅对已经表现出可疑行为的访问者发起挑战，最大限度减少对正常用户的影响。

## 跳过挑战的情况

以下请求会自动跳过 JS 挑战：

- **白名单 IP** -- 白名单中的 IP 始终免于挑战
- **已认证请求** -- 携带有效 `Defender-Authorization` 请求头的请求
- **Git/许可证令牌** -- 携带配置的 Git 或许可证令牌请求头的请求
- **非浏览器客户端** -- 被识别为 `git`、`curl`、`wget` 等的客户端
- **Auth 子请求** -- Nginx `auth_request` 使用的 `/auth` 端点

## 难度设置

难度设置控制解决挑战所需的计算量：

| 难度 | 前导零数 | 大约迭代次数 |
|------|---------|------------|
| 1 | 1 | ~16 |
| 2 | 2 | ~256 |
| 3 | 3 | ~4,096 |
| **4**（默认） | **4** | **~65,536** |
| 5 | 5 | ~1,048,576 |
| 6 | 6 | ~16,777,216 |

难度越高，客户端需要的计算时间越长。默认值 4 在机器人防护和用户体验之间取得了良好平衡（在现代设备上通常不到 2 秒即可完成）。

## 配置

```yaml
js-challenge:
  enabled: false
  # 模式：off | suspicious | all
  mode: "suspicious"
  # 难度：SHA256 哈希前导零数量（1-6）
  difficulty: 4
  # 通行 Cookie 有效期（秒），默认 24 小时
  cookie-ttl: 86400
  # Cookie 签名密钥（留空则自动生成）
  cookie-secret: ""
```

JS 挑战设置也可以通过管理后台的**系统设置**页面进行配置。

!!! warning "生产环境 Cookie 密钥"
    如果 `cookie-secret` 留空，每次重启时会生成随机密钥，导致所有已有的通行 Cookie 失效。请在生产环境中设置固定密钥。

---

## 相关页面

- [威胁检测](threat-detection.md) -- 威胁评分如何驱动 `suspicious` 模式
- [安全事件](security-events.md) -- JS 挑战失败会被记录为安全事件
- [配置说明](../configuration/index.md) -- 完整配置参考
