# Quick Start: Row Level Security Resource

## Installation

1. Build the provider:
```bash
go install .
```

2. The provider will be installed to `$GOPATH/bin`

## Basic Usage

### 1. Configure the Provider

```terraform
terraform {
  required_providers {
    superset = {
      source = "svdimchenko/superset"
    }
  }
}

provider "superset" {
  host     = "http://localhost:8088"
  username = "admin"
  password = "admin"
}
```

### 2. Create an RLS Rule

```terraform
resource "superset_row_level_security" "my_rls" {
  table_id    = 1
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [1, 2]
  group_key   = "user_filter"
  filter_type = "Regular"  # Optional: 'Regular' (default) or 'Base'
}
```

### 3. Apply the Configuration

```bash
terraform init
terraform plan
terraform apply
```

## Import Existing RLS Rules

```bash
terraform import superset_row_level_security.my_rls 123
```

Where `123` is the ID of the existing RLS rule in Superset.

## Common Patterns

### User-based Filtering
```terraform
resource "superset_row_level_security" "user_data" {
  table_id    = var.table_id
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [var.analyst_role_id]
  filter_type = "Regular"
}
```

### Department-based Filtering
```terraform
resource "superset_row_level_security" "dept_data" {
  table_id    = var.table_id
  clause      = "department = '{{ current_user.department }}'"
  role_ids    = [var.manager_role_id]
  group_key   = "department"
  filter_type = "Regular"
}
```

### Base Filter (Applied to All)
```terraform
resource "superset_row_level_security" "base_filter" {
  table_id    = var.table_id
  clause      = "status = 'active'"
  filter_type = "Base"
}
```

### Time-based Filtering
```terraform
resource "superset_row_level_security" "recent_data" {
  table_id    = var.table_id
  clause      = "created_at >= CURRENT_DATE - INTERVAL '90 days'"
  role_ids    = [var.viewer_role_id]
  filter_type = "Regular"
}
```

## Troubleshooting

### Check RLS Rule Status
```bash
terraform show
```

### Refresh State
```bash
terraform refresh
```

### Debug Mode
```bash
TF_LOG=DEBUG terraform apply
```

## Next Steps

- Review the [full documentation](../../docs/resources/row_level_security.md)
- Check out [more examples](README.md)
- Read about [Superset RLS](https://superset.apache.org/docs/security/#row-level-security)
