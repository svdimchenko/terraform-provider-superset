# Example: Using LDAP authentication provider
provider "superset" {
  host     = "https://domain.com" # Replace with your Superset instance URL
  username = "ldap_user"          # Replace with your LDAP username
  password = "ldap_password"      # Replace with your LDAP password
  provider = "ldap"               # Use LDAP authentication
}

# Alternative: Using environment variables
# export SUPERSET_HOST="https://domain.com"
# export SUPERSET_USERNAME="ldap_user"
# export SUPERSET_PASSWORD="ldap_password"
# export SUPERSET_PROVIDER="ldap"
#
# provider "superset" {}
