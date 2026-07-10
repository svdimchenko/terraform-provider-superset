resource "superset_chart_import" "example" {
  source_dir      = "${path.module}/dashboards/athena_usage"
  force_overwrite = true

  database_secrets = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = var.athena_db_password
  }

  database_overrides = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = jsonencode({
      sqlalchemy_uri = "awsathena+rest://athena.eu-central-1.amazonaws.com/?s3_staging_dir=s3%3A%2F%2Ftest-athena%2Fresults%2Fsuperset&work_group=test"
    })
  }
}
