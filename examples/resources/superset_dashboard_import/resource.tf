resource "superset_dashboard_import" "example" {
  source_dir      = "${path.module}/dashboards/athena_usage"
  force_overwrite = true

  database_secrets = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = var.athena_db_password
  }

  # Override any database YAML fields (including nested) by UUID.
  # Value is a JSON-encoded object that gets deep-merged into the YAML.
  database_overrides = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = jsonencode({
      sqlalchemy_uri = "awsathena+rest://athena.eu-central-1.amazonaws.com/?s3_staging_dir=s3%3A%2F%2Fstage-dwh-athena%2Fresults%2Fsuperset&work_group=superset"
      extra = {
        schemas_allowed_for_file_upload = ["dev_sandbox"]
        cost_estimate_enabled           = false
      }
    })
  }
}
