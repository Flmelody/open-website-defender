# Database Configuration

Website Defender supports three database backends. Choose the one that best fits your deployment requirements.

## Supported Databases

| Database | Driver Name | Default Connection | Best For |
|----------|------------|-------------------|----------|
| **SQLite** | `sqlite` | `./data/app.db` | Single-node deployments, development, small-to-medium workloads |
| **PostgreSQL** | `postgres` | `localhost:5432` | Production deployments, high concurrency |
| **MySQL** | `mysql` | `localhost:3306` | Production deployments, existing MySQL infrastructure |

!!! info "Default Database"
    SQLite is the default database and requires no additional setup. The database file is created automatically on first startup.

## Configuration Examples

=== "SQLite"

    SQLite is the simplest option -- no external database server required. The database is stored as a single file.

    ```yaml
    database:
      driver: sqlite
      # Optional: customize the database file path
      # file-path: ./data/app.db
    ```

    !!! tip "SQLite File Location"
        The default path is `./data/app.db` relative to the working directory. Make sure the directory exists and is writable, or specify an absolute path.

=== "PostgreSQL"

    PostgreSQL is recommended for production deployments with high concurrency requirements.

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

    Create the database before starting Website Defender:

    ```sql
    CREATE DATABASE open_website_defender;
    ```

    !!! warning "SSL Mode"
        In production, set `ssl-mode` to `require` or `verify-full` to encrypt database connections. Only use `disable` for local development.

=== "MySQL"

    MySQL is a good choice if you have existing MySQL infrastructure.

    ```yaml
    database:
      driver: mysql
      host: localhost
      port: 3306
      name: open_website_defender
      user: root
      password: your_password
    ```

    Create the database before starting Website Defender:

    ```sql
    CREATE DATABASE open_website_defender CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    ```

    !!! tip "Character Set"
        Use `utf8mb4` character set to ensure full Unicode support, especially if usernames or descriptions contain non-ASCII characters.

## Switching Databases

To switch from SQLite to PostgreSQL or MySQL:

1. Update `config/config.yaml` with the new driver and connection settings
2. Create the target database (if using PostgreSQL or MySQL)
3. Restart Website Defender -- tables are created automatically on startup
4. Recreate your users, IP lists, and WAF rules in the new database

!!! warning "No Automatic Migration"
    Switching databases does not migrate existing data. You will need to recreate your configuration in the new database, or manually export and import the data.
