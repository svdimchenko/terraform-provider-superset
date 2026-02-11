# Row Level Security Resource Example

This example demonstrates how to create a row level security (RLS) rule in Apache Superset using Terraform.

## What is Row Level Security?

Row Level Security (RLS) in Superset allows you to filter data at the row level based on user roles. This ensures that users only see the data they are authorized to access.

## Usage

```terraform
# Fetch roles by name
data "superset_role" "analyst" {
  name = "Analyst"
}

resource "superset_row_level_security" "example" {
  tables      = [1]
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [data.superset_role.analyst.id]
  group_key   = "department"
  filter_type = "Regular"
  description = "Filter data by current user ID"
}
```

## Arguments

- `tables` (Required) - List of table/dataset IDs to apply the RLS rule to
- `clause` (Required) - SQL WHERE clause that will be applied to filter rows. Can use Jinja templates like `{{ current_user_id() }}`
- `role_ids` (Optional) - List of role IDs that this RLS rule applies to
- `group_key` (Optional) - Group key for organizing related RLS rules
- `filter_type` (Optional) - Filter type: 'Regular' or 'Base'. Defaults to 'Regular'
- `description` (Optional) - Description of the RLS rule

## Common Use Cases

### Filter by Current User
```terraform
data "superset_role" "analyst" {
  name = "Analyst"
}

resource "superset_row_level_security" "user_filter" {
  tables      = [1]
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [data.superset_role.analyst.id]
  filter_type = "Regular"
  description = "Restrict analysts to their own data"
}
```

### Filter by Department
```terraform
data "superset_role" "manager" {
  name = "Manager"
}

data "superset_role" "director" {
  name = "Director"
}

resource "superset_row_level_security" "department_filter" {
  tables      = [2]
  clause      = "department = '{{ current_user.department }}'"
  role_ids    = [data.superset_role.manager.id, data.superset_role.director.id]
  group_key   = "department"
  filter_type = "Regular"
  description = "Filter by user's department for managers and directors"
}
```

### Base Filter (Applied to All Users)
```terraform
resource "superset_row_level_security" "base_filter" {
  tables      = [3]
  clause      = "status = 'active'"
  filter_type = "Base"
  description = "Only show active records to all users"
}
```

### Filter by Date Range
```terraform
resource "superset_row_level_security" "date_filter" {
  tables      = [4]
  clause      = "created_at >= CURRENT_DATE - INTERVAL '30 days'"
  role_ids    = [4]
  filter_type = "Regular"
}
```
