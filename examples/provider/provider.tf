# Basic provider configuration
# See provider-db.tf and provider-ldap.tf for specific authentication examples

provider "superset" {
  host     = "https://domain.com" # Replace with your Superset instance URL
  username = "username"           # Replace with your Superset username
  password = "password"           # Replace with your Superset password
  provider = "db"                 # Authentication provider: "db" (default) or "ldap"
}
