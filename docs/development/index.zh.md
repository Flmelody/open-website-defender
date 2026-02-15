# 开发指南

本页面介绍如何从源码构建和开发 Website Defender。

## 环境要求

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| **Go** | 1.25+ | 后端编译和开发 |
| **Node.js** | 20+ | 前端构建和开发 |
| **npm** | 随 Node.js 安装 | 前端包管理 |

## 从源码构建

### 完整构建

使用构建脚本一键完成前端和后端的编译：

```bash
# 克隆仓库
git clone https://github.com/Flmelody/open-website-defender.git
cd open-website-defender

# 安装前端依赖
cd ui/admin && npm install && cd ../..
cd ui/guard && npm install && cd ../..

# 执行构建脚本
./scripts/build.sh
```

构建完成后，根目录下会生成 `app` 可执行文件。

### 分步构建

如需分别构建前端和后端：

**构建前端：**

```bash
# 构建 Guard 防护页
cd ui/guard
npm install
npm run build
cd ../..

# 构建 Admin 管理后台
cd ui/admin
npm install
npm run build
cd ../..
```

**构建后端：**

!!! warning "前端依赖"
    Go 后端通过 `go:embed` 嵌入 `ui/admin/dist` 和 `ui/guard/dist` 目录。**必须先构建前端**，否则后端编译会失败。

```bash
go build -o app main.go
```

## 开发服务器

### 后端开发

```bash
# 运行开发服务器（默认端口 9999）
go run main.go

# 使用自定义配置
go run main.go -config ./config/config.yaml
```

### 前端开发

项目包含两个独立的 Vue 3 前端应用：

```bash
# Admin 管理后台开发服务器
cd ui/admin
npm run dev

# Guard 防护页开发服务器
cd ui/guard
npm run dev
```

!!! tip "前端开发环境"
    前端开发服务器默认连接 `http://localhost:9999/wall` 后端 API。如需修改，可通过环境变量 `VITE_BACKEND_HOST` 配置。

### 前端代码检查

```bash
# 类型检查
cd ui/admin && npm run type-check
cd ui/guard && npm run type-check

# 代码风格检查和修复
cd ui/admin && npm run lint
cd ui/guard && npm run lint
```

## 项目结构

```
open-website-defender/
├── main.go                    # 应用入口
├── config/
│   └── config.yaml            # 运行时配置文件
├── scripts/
│   └── build.sh               # 构建脚本
├── internal/
│   ├── adapter/
│   │   ├── controller/http/   # Gin HTTP 处理器、中间件、请求/响应结构
│   │   └── repository/        # GORM 数据库仓储实现
│   ├── domain/
│   │   ├── entity/            # 领域模型（User, IpBlackList, IpWhiteList 等）
│   │   └── error/             # 领域错误定义
│   ├── infrastructure/
│   │   ├── config/            # 配置结构体
│   │   ├── database/          # 数据库初始化（SQLite/PostgreSQL/MySQL via GORM）
│   │   └── logging/           # Zap 日志初始化
│   ├── pkg/                   # 工具包（JWT、加密、HTTP 辅助函数）
│   └── usecase/               # 业务逻辑服务层（含 DTO）
│       ├── interface/         # 仓储接口定义
│       ├── iplist/            # IP 名单服务
│       └── user/              # 认证和用户服务
├── ui/
│   ├── admin/                 # Admin 管理后台（Vue 3 + Element Plus + vue-i18n）
│   └── guard/                 # Guard 防护页（Vue 3）
└── docs/                      # MkDocs 文档
```

### 关键设计模式

- **Clean Architecture（整洁架构）**：代码分层为 adapter、domain、infrastructure、usecase
- **单例服务**：使用 `sync.Once` 模式初始化服务（如 `GetAuthService()`、`GetIpBlackListService()`）
- **仓储接口**：定义在 `internal/usecase/interface/repository.go`，实现在 `internal/adapter/repository/`
- **JWT 认证**：令牌通过 `Defender-Authorization` 请求头或 `flmelody.token` Cookie 传递

## 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并输出详细信息
go test -v ./...

# 运行特定包的测试
go test ./internal/usecase/...
```

---

## 相关页面

- [快速开始](../getting-started/index.md) - 快速体验指南
- [架构说明](../architecture/index.md) - 系统架构详解
- [环境变量](../configuration/environment-variables.md) - 构建时环境变量
