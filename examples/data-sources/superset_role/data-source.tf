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

# Fetch a role by name
data "superset_role" "analyst" {
  name = "Analyst"
}

# Use the role ID in a resource
resource "superset_row_level_security" "example" {
  table_id    = 1
  clause      = "department = 'sales'"
  role_ids    = [data.superset_role.analyst.id]
  filter_type = "Regular"
}

# Output the role ID
output "analyst_role_id" {
  value = data.superset_role.analyst.id
}
