resource "superset_dataset_import" "all" {
  source_dir      = "${path.module}/dashboards"
  force_overwrite = true

  database_secrets = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = var.db_password
  }

  database_overrides = {
    "dd568dff-7835-4cee-8e42-c91f3b533c49" = jsonencode({
      sqlalchemy_uri = "starrocks://admin@starrocks-host:9030"
    })
  }
}
