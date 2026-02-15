# Geo-IP Blocking

Website Defender can block requests based on the geographic location of the client's IP address using the MaxMind GeoLite2-Country database.

## How It Works

When enabled, every incoming request is checked against the GeoLite2-Country database to determine the client's country. If the country code matches a blocked country in the configured rules, the request is denied with a `403 Forbidden` response.

Geo-blocking is applied in the [middleware chain](../architecture/index.md) **before** the WAF and rate limiter, ensuring blocked countries are rejected early in the request pipeline.

!!! info "MaxMind GeoLite2 Database Required"
    Geo-blocking requires the MaxMind GeoLite2-Country database file (`.mmdb` format). You can download it for free from [MaxMind's website](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) after creating an account.

## Configuration

Enable geo-blocking and specify the path to your GeoLite2-Country database in `config/config.yaml`:

```yaml
geo-blocking:
  enabled: true
  database-path: "/path/to/GeoLite2-Country.mmdb"
```

| Setting | Description | Default |
|---------|-------------|---------|
| `enabled` | Enable or disable geo-blocking | `false` |
| `database-path` | Path to the MaxMind GeoLite2-Country `.mmdb` file | `""` |

For the full configuration reference, see [Configuration](../configuration/index.md).

## Managing Blocked Countries

Blocked country codes are managed via the **admin dashboard** or the API:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/geo-block-rules` | List all blocked country codes |
| `POST` | `/geo-block-rules` | Add a country code to the block list |
| `DELETE` | `/geo-block-rules/:id` | Remove a country code from the block list |

!!! tip "ISO 3166-1 Country Codes"
    Country codes follow the ISO 3166-1 alpha-2 standard (e.g., `CN` for China, `RU` for Russia, `US` for United States). Use two-letter uppercase codes when adding blocked countries.

## Considerations

- **Database updates**: The GeoLite2 database is updated regularly by MaxMind. Consider setting up a periodic download to keep your country mappings current.
- **Performance**: The MMDB lookup is highly efficient (in-memory B-tree), adding negligible latency to request processing.
- **Proxy IPs**: If your server is behind a proxy, ensure that the real client IP is being passed correctly via `X-Forwarded-For`. See [Nginx Setup](../deployment/nginx-setup.md) for configuration details.
