package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jarcoal/httpmock"
)

func TestAccRolePermissionsResource(t *testing.T) {

	t.Run("CreateRead", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "http://superset-host/api/v1/security/login",
			httpmock.NewStringResponder(200, `{"access_token": "fake-token"}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles/129",
			httpmock.NewStringResponder(200, `{"result": {"id": 129, "name": "DWH-DB-Connect"}}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles?q=(page_size:5000)",
			httpmock.NewStringResponder(200, `{
				"result": [
					{"id": 129, "name": "DWH-DB-Connect"}
				]
			}`))

		// Mock paginated permissions-resources (used by GetPermissionIDByNameAndView)
		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/permissions-resources?q=(page:0,page_size:100)",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission": {
						"name": "database_access"
					},
					"view_menu": {
						"name": "[SelfPostgreSQL].(id:1)"
					}
				}
			]}`))

		// Also mock the old-style URL with trailing slash (used by GetPermissionIDs)
		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/permissions-resources/?q=(page:0,page_size:100)",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission": {
						"name": "database_access"
					},
					"view_menu": {
						"name": "[SelfPostgreSQL].(id:1)"
					}
				}
			]}`))

		httpmock.RegisterResponder("POST", "http://superset-host/api/v1/security/roles/129/permissions",
			httpmock.NewStringResponder(200, `{"status": "success"}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles/129/permissions/",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission_name": "database_access",
					"view_menu_name": "[SelfPostgreSQL].(id:1)"
				}
			]}`))

		httpmock.RegisterResponder("DELETE", "http://superset-host/api/v1/security/roles/129/permissions",
			httpmock.NewStringResponder(204, ""))

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: providerConfig + `
resource "superset_role_permissions" "team" {
  role_name            = "DWH-DB-Connect"
  resource_permissions = [
    {
      permission = "database_access"
      view_menu  = "[SelfPostgreSQL].(id:1)"
    }
  ]
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("superset_role_permissions.team", "role_name", "DWH-DB-Connect"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.#", "1"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.0.permission", "database_access"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.0.view_menu", "[SelfPostgreSQL].(id:1)"),
					),
				},
			},
		})
	})

	t.Run("UpdateRead", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", "http://superset-host/api/v1/security/login",
			httpmock.NewStringResponder(200, `{"access_token": "fake-token"}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles?q=(page_size:5000)",
			httpmock.NewStringResponder(200, `{
				"result": [
					{"id": 129, "name": "DWH-DB-Connect"}
				]
			}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles/129",
			httpmock.NewStringResponder(200, `{"result": {"id": 129, "name": "DWH-DB-Connect"}}`))

		// Mock paginated permissions-resources
		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/permissions-resources?q=(page:0,page_size:100)",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission": {
						"name": "database_access"
					},
					"view_menu": {
						"name": "[SelfPostgreSQL].(id:1)"
					}
				},
				{
					"id": 241,
					"permission": {
						"name": "schema_access"
					},
					"view_menu": {
						"name": "[Trino].[devoriginationzestorage]"
					}
				}
			]}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/permissions-resources/?q=(page:0,page_size:100)",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission": {
						"name": "database_access"
					},
					"view_menu": {
						"name": "[SelfPostgreSQL].(id:1)"
					}
				},
				{
					"id": 241,
					"permission": {
						"name": "schema_access"
					},
					"view_menu": {
						"name": "[Trino].[devoriginationzestorage]"
					}
				}
			]}`))

		httpmock.RegisterResponder("POST", "http://superset-host/api/v1/security/roles/129/permissions",
			httpmock.NewStringResponder(200, `{"status": "success"}`))

		httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/roles/129/permissions/",
			httpmock.NewStringResponder(200, `{ "result": [
				{
					"id": 240,
					"permission_name": "database_access",
					"view_menu_name": "[SelfPostgreSQL].(id:1)"
				},
				{
					"id": 241,
					"permission_name": "schema_access",
					"view_menu_name": "[Trino].[devoriginationzestorage]"
				}
			]}`))

		httpmock.RegisterResponder("DELETE", "http://superset-host/api/v1/security/roles/129/permissions",
			httpmock.NewStringResponder(204, ""))

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: providerConfig + `
resource "superset_role_permissions" "team" {
  role_name            = "DWH-DB-Connect"
  resource_permissions = [
    {
      permission = "database_access"
      view_menu  = "[SelfPostgreSQL].(id:1)"
    },
    {
      permission = "schema_access"
      view_menu  = "[Trino].[devoriginationzestorage]"
    },
  ]
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("superset_role_permissions.team", "role_name", "DWH-DB-Connect"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.#", "2"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.1.permission", "schema_access"),
						resource.TestCheckResourceAttr("superset_role_permissions.team", "resource_permissions.1.view_menu", "[Trino].[devoriginationzestorage]"),
					),
				},
			},
		})
	})
}
