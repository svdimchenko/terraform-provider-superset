resource "superset_dashboard_embedding" "example" {
  dashboard_id    = superset_dashboard_import.example.dashboard_id
  allowed_domains = ["https://app.example.com", "https://portal.example.com"]
}

output "embedded_dashboard_uuid" {
  value = superset_dashboard_embedding.example.uuid
}
