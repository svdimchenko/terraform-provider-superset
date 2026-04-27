resource "superset_dashboard_import" "example" {
  source_dir      = "${path.module}/dashboards/my_dashboard"
  force_overwrite = true

  database_secrets = {
    "<uuid>" = var.database_password
  }
}
