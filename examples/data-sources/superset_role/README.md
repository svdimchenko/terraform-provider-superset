# Role Data Source Example

This example demonstrates how to fetch a role by name from Apache Superset using Terraform.

## What is the Role Data Source?

The `superset_role` data source allows you to look up a role by its name and retrieve its ID.
This is useful when you need to reference roles in other resources without hardcoding role IDs.

## Basic Usage

```terraform
data "superset_role" "analyst" {
  name = "Analyst"
}

output "analyst_role_id" {
  value = data.superset_role.analyst.id
}
```

## Use Cases

### 1. Reference Role in RLS Rules

```terraform
data "superset_role" "analyst" {
  name = "Analyst"
}

resource "superset_row_level_security" "analyst_filter" {
  table_id    = 1
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [data.superset_role.analyst.id]
  filter_type = "Regular"
}
```

### 2. Multiple Roles

```terraform
data "superset_role" "analyst" {
  name = "Analyst"
}

data "superset_role" "manager" {
  name = "Manager"
}

resource "superset_row_level_security" "multi_role_filter" {
  table_id    = 1
  clause      = "department = 'sales'"
  role_ids    = [
    data.superset_role.analyst.id,
    data.superset_role.manager.id
  ]
  filter_type = "Regular"
}
```

### 3. Dynamic Role Assignment

```terraform
variable "role_names" {
  type    = list(string)
  default = ["Analyst", "Manager", "Director"]
}

data "superset_role" "roles" {
  for_each = toset(var.role_names)
  name     = each.value
}

resource "superset_row_level_security" "dynamic_filter" {
  table_id    = 1
  clause      = "status = 'active'"
  role_ids    = [for role in data.superset_role.roles : role.id]
  filter_type = "Regular"
}
```

## Attributes

### Input

- `name` (Required) - The name of the role to fetch

### Output

- `id` (Computed) - The numeric identifier of the role

## Common Role Names

Superset typically includes these default roles:

- `Admin` - Full administrative access
- `Alpha` - Can create and edit content
- `Gamma` - Can view dashboards and charts
- `Public` - Limited public access
- `sql_lab` - SQL Lab access

Custom roles can also be fetched by their name.
