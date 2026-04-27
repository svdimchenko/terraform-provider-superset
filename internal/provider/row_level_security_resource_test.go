package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jarcoal/httpmock"
)

func TestAccRowLevelSecurityResource(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock login
	httpmock.RegisterResponder("POST", "http://superset-host/api/v1/security/login",
		httpmock.NewStringResponder(200, `{"access_token": "fake-token"}`))

	// Mock CSRF token
	httpmock.RegisterResponder("GET", "http://superset-host/api/v1/security/csrf_token/",
		httpmock.NewStringResponder(200, `{"result": "fake-csrf-token"}`))

	// Mock create RLS
	httpmock.RegisterResponder("POST", "http://superset-host/api/v1/rowlevelsecurity/",
		httpmock.NewStringResponder(201, `{"id": 1}`))

	// Mock read RLS
	httpmock.RegisterResponder("GET", "http://superset-host/api/v1/rowlevelsecurity/1",
		httpmock.NewStringResponder(200, `{
			"result": {
				"id": 1,
				"name": "test_rls_rule",
				"tables": [{"id": 1}],
				"clause": "user_id = '{{ current_user_id() }}'",
				"roles": [{"id": 1}],
				"group_key": "test",
				"filter_type": "Regular",
				"description": "Test RLS rule"
			}
		}`))

	// Mock update RLS
	httpmock.RegisterResponder("PUT", "http://superset-host/api/v1/rowlevelsecurity/1",
		httpmock.NewStringResponder(200, `{}`))

	// Mock delete RLS
	httpmock.RegisterResponder("DELETE", "http://superset-host/api/v1/rowlevelsecurity/1",
		httpmock.NewStringResponder(200, `{}`))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "superset_row_level_security" "test" {
  name        = "test_rls_rule"
  tables      = [1]
  clause      = "user_id = '{{ current_user_id() }}'"
  role_ids    = [1]
  group_key   = "test"
  filter_type = "Regular"
  description = "Test RLS rule"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("superset_row_level_security.test", "name", "test_rls_rule"),
					resource.TestCheckResourceAttr("superset_row_level_security.test", "clause", "user_id = '{{ current_user_id() }}'"),
					resource.TestCheckResourceAttr("superset_row_level_security.test", "group_key", "test"),
					resource.TestCheckResourceAttr("superset_row_level_security.test", "filter_type", "Regular"),
					resource.TestCheckResourceAttr("superset_row_level_security.test", "description", "Test RLS rule"),
				),
			},
			{
				ResourceName:      "superset_row_level_security.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
