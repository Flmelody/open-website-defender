# 快速开始

本页面将指导您完成 Castellum 的环境准备、构建和运行。

## 环境要求

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| **Go** | 1.25+ | 后端编译 |
| **Node.js** | 20+ | 前端构建 |
| **Nginx** | 任意版本 | 需包含 `auth_request` 模块 |

!!! info "关于 Nginx"
    Nginx 的 `auth_request` 模块是 Castellum 的核心依赖。该模块允许 Nginx 在转发请求之前，先向 Castellum 发起子请求以验证用户身份。大多数主流 Nginx 安装包已默认包含此模块。

## 构建

项目包含一个构建脚本，用于编译前端和后端代码。

```bash
# 1. 克隆仓库
git clone https://github.com/Flmelody/castellum.git
cd castellum

# 2. 构建项目
./scripts/build.sh
```

!!! tip "自定义构建配置"
    您可以通过修改 `scripts/build.sh` 或在项目根目录创建 `.env` 文件来自定义构建配置。支持的环境变量请参阅 [环境变量](../configuration/environment-variables.md) 页面。

构建脚本会依次完成以下步骤：

1. 构建 Guard 前端（`ui/guard`）
2. 构建 Admin 前端（`ui/admin`）
3. 编译 Go 后端并嵌入前端资源

构建完成后，根目录下会生成一个名为 `app` 的可执行文件。

## 运行

```bash
# 运行应用
./app
```

应用将使用默认配置启动：

| 项目 | 值 |
|------|------|
| 管理后台地址 | `http://localhost:9999/wall/admin/` |
| 默认用户名 | `castellum` |
| 默认密码 | 首次启动自动生成，写入 `./data/bootstrap-admin-credentials`（权限 `0600`） |

!!! warning "安全警告"
    首次启动时默认密码会随机生成并写入 `./data/bootstrap-admin-credentials`（权限 `0600`）。启动日志只会打印该文件路径，请打开该文件读取密码。完成首次登录并**轮换密码后请删除该文件**。如希望跳过自动生成，可在 `config/config.yaml` 中配置 `default-user.username` / `default-user.password`。

!!! warning "从 Open Website Defender 升级（PostgreSQL / MySQL）"
    内置默认数据库名从 `open_website_defender` 改为 `castellum`。如果你的 `config/config.yaml` **没有显式配置** `database.name` 且正在从老版本升级 PG/MySQL 部署，升级前请显式写上 `database.name: open_website_defender`，或者把数据库 rename 成 `castellum`，否则连接的会是一个新的/不存在的库。SQLite 部署不受影响。

!!! tip "下一步"
    - 配置 Nginx 以集成 Castellum，请参阅 [Nginx 配置](../deployment/nginx-setup.md)
    - 了解运行时配置选项，请参阅 [配置说明](../configuration/index.md)
    - 查看完整功能列表，请参阅 [认证与访问控制](../features/authentication.md)
