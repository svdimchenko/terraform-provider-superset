package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRowLevelSecurityResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
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
