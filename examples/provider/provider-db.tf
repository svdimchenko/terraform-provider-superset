# Example: Using database authentication provider (default)
provider "superset" {
  host     = "https://domain.com" # Replace with your Superset instance URL
  username = "admin"              # Replace with your database username
  password = "admin_password"     # Replace with your database password
  provider = "db"                 # Use database authentication (this is the default)
}

# Alternative: Omit provider field (defaults to "db")
# provider "superset" {
#   host     = "https://domain.com"
#   username = "admin"
#   password = "admin_password"
# }

# Alternative: Using environment variables
# export SUPERSET_HOST="https://domain.com"
# export SUPERSET_USERNAME="admin"
# export SUPERSET_PASSWORD="admin_password"
# export SUPERSET_PROVIDER="db"  # Optional, defaults to "db"
#
# provider "superset" {}
