# 快速开始

本页面将指导您完成 Website Defender 的环境准备、构建和运行。

## 环境要求

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| **Go** | 1.25+ | 后端编译 |
| **Node.js** | 20+ | 前端构建 |
| **Nginx** | 任意版本 | 需包含 `auth_request` 模块 |

!!! info "关于 Nginx"
    Nginx 的 `auth_request` 模块是 Website Defender 的核心依赖。该模块允许 Nginx 在转发请求之前，先向 Defender 发起子请求以验证用户身份。大多数主流 Nginx 安装包已默认包含此模块。

## 构建

项目包含一个构建脚本，用于编译前端和后端代码。

```bash
# 1. 克隆仓库
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender

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
| 默认用户名 | `defender` |
| 默认密码 | `defender` |

!!! warning "安全警告"
    默认凭据 `defender / defender` 仅用于初始设置。**请在首次登录后立即修改密码**，避免在生产环境中使用默认凭据。

!!! tip "下一步"
    - 配置 Nginx 以集成 Defender，请参阅 [Nginx 配置](../deployment/nginx-setup.md)
    - 了解运行时配置选项，请参阅 [配置说明](../configuration/index.md)
    - 查看完整功能列表，请参阅 [认证与访问控制](../features/authentication.md)
