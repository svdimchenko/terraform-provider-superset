# Provider Configuration Examples

This directory contains examples of how to configure the Superset provider with different authentication methods.

## Authentication Providers

The Superset provider supports two authentication providers:

### 1. Database Authentication (default)

Use `provider = "db"` or omit the provider field entirely to authenticate using Superset's internal database.

**Example:**
```terraform
provider "superset" {
  host     = "https://superset.example.com"
  username = "admin"
  password = "admin_password"
  provider = "db"  # Optional, this is the default
}
```

See [provider-db.tf](provider-db.tf) for a complete example.

### 2. LDAP Authentication

Use `provider = "ldap"` to authenticate using LDAP/Active Directory.

**Example:**
```terraform
provider "superset" {
  host     = "https://superset.example.com"
  username = "ldap_user"
  password = "ldap_password"
  provider = "ldap"
}
```

See [provider-ldap.tf](provider-ldap.tf) for a complete example.

## Configuration Options

| Attribute | Description | Required | Default |
|-----------|-------------|----------|---------|
| `host` | The URL of the Superset instance (including protocol) | Yes | - |
| `username` | Username for authentication | Yes | - |
| `password` | Password for authentication | Yes | - |
| `provider` | Authentication provider: `db` or `ldap` | No | `db` |

## Environment Variables

All configuration options can be set via environment variables:

- `SUPERSET_HOST` - Superset instance URL
- `SUPERSET_USERNAME` - Username for authentication
- `SUPERSET_PASSWORD` - Password for authentication
- `SUPERSET_PROVIDER` - Authentication provider (`db` or `ldap`)

**Example:**
```bash
export SUPERSET_HOST="https://superset.example.com"
export SUPERSET_USERNAME="admin"
export SUPERSET_PASSWORD="admin_password"
export SUPERSET_PROVIDER="ldap"

terraform plan
```

When using environment variables, you can use an empty provider block:
```terraform
provider "superset" {}
```

## Notes

- The provider field is case-sensitive. Use lowercase `db` or `ldap`.
- If the provider field is not specified, it defaults to `db`.
- Environment variables are overridden by explicit configuration values in the provider block.
