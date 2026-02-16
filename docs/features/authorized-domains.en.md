# Authorized Domains

Authorized Domains is a centralized registry of all protected domains managed by Website Defender. It serves as the **single source of truth** for domain names used across the system -- in IP whitelist bindings and user access scopes.

## Purpose

Without Authorized Domains, users must manually type domain names when configuring IP whitelist entries or user scopes, which is error-prone and inconsistent. Authorized Domains solves this by:

- Providing a **central place** to register all protected domains
- Populating **dropdown selectors** in the IP whitelist and user management forms
- Ensuring **cascade cleanup** when a domain is removed

## Management

### Admin Dashboard

Manage authorized domains through the admin dashboard:

1. Navigate to the **Authorized Domains** page
2. Add domains by entering the domain name (e.g., `gitea.example.com`)
3. View all registered domains with creation timestamps
4. Delete domains when they are no longer needed

### API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/authorized-domains` | List authorized domains (paginated) |
| `GET` | `/authorized-domains?all=true` | List all authorized domains (for dropdowns) |
| `POST` | `/authorized-domains` | Register a new authorized domain |
| `DELETE` | `/authorized-domains/:id` | Remove an authorized domain |

!!! note "Duplicate Prevention"
    Domain names must be unique. Attempting to register a domain that already exists will return a `409 Conflict` response.

## Cascade Deletion

When an authorized domain is deleted, the following cleanup happens automatically:

1. **IP Whitelist**: All whitelist entries bound to this domain are removed
2. **User Scopes**: This domain is removed from all users' scope lists

!!! warning "Cascade Effects"
    Deleting an authorized domain may remove IP whitelist entries and change user access permissions. Review the impact before deleting a domain that is actively in use.

## Integration with Other Features

### IP Whitelist

When adding an IP whitelist entry, the **Authorized Domain** field provides a dropdown populated from the authorized domains registry. You can also type a custom value if needed.

See [IP Lists](ip-lists.md) for details.

### User Scopes

When configuring user access scopes, the **Authorized Domains** field provides a multi-select dropdown populated from the registry. Users can select multiple domains or type custom patterns (e.g., `*.example.com`).

See [User Management](user-management.md) and [Domain Scopes](domain-scopes.md) for details.

---

## Related Pages

- [IP Lists](ip-lists.md) -- IP whitelist domain binding
- [User Management](user-management.md) -- User scope assignment
- [Domain Scopes](domain-scopes.md) -- How domain scope matching works
- [API Reference](../api-reference/index.md) -- Full API documentation
