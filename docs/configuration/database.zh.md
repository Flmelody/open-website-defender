# 数据库

Website Defender 支持多种数据库后端，可根据部署环境和性能需求选择合适的数据库。

## 支持的数据库

| 数据库 | 驱动名称 | 默认配置 | 适用场景 |
|--------|---------|----------|---------|
| **SQLite** | `sqlite` | `./data/app.db` | 单机部署、轻量使用、快速体验 |
| **PostgreSQL** | `postgres` | `localhost:5432` | 生产环境、高并发、数据可靠性要求高 |
| **MySQL** | `mysql` | `localhost:3306` | 生产环境、团队熟悉 MySQL 生态 |

!!! info "默认数据库"
    如果不指定数据库配置，Website Defender 默认使用 SQLite，数据文件存储在 `./data/app.db`。SQLite 无需额外安装数据库服务，适合快速体验和小规模部署。

## 配置示例

在 `config/config.yaml` 中配置数据库连接：

=== "SQLite"

    ```yaml
    database:
      driver: sqlite
      # 数据库文件路径（可选，默认 ./data/app.db）
      # file-path: ./data/app.db
    ```

    !!! tip "SQLite 注意事项"
        - 数据文件会自动创建
        - 适合单实例部署
        - 不支持多进程并发写入
        - 建议定期备份数据文件

=== "PostgreSQL"

    ```yaml
    database:
      driver: postgres
      host: localhost
      port: 5432
      name: open_website_defender
      user: postgres
      password: your_password
      ssl-mode: disable
    ```

    !!! tip "PostgreSQL 注意事项"
        - 需要预先创建数据库（如 `open_website_defender`）
        - 生产环境建议启用 SSL（`ssl-mode: require`）
        - 支持高并发访问
        - 建议配置连接池参数

=== "MySQL"

    ```yaml
    database:
      driver: mysql
      host: localhost
      port: 3306
      name: open_website_defender
      user: root
      password: your_password
    ```

    !!! tip "MySQL 注意事项"
        - 需要预先创建数据库（如 `open_website_defender`）
        - 建议使用 `utf8mb4` 字符集
        - 建议为 Defender 创建独立的数据库用户，避免使用 `root`

!!! warning "密码安全"
    不要将数据库密码直接写入配置文件并提交到版本控制系统。建议通过环境变量或密钥管理服务注入敏感配置。

---

## 相关页面

- [配置说明](index.md) - 完整的运行时配置参考
- [部署指南](../deployment/index.md) - 部署建议
