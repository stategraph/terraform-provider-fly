package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccAppsDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAppsDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_apps.test", "apps.#", 1),
				),
			},
		},
	})
}

func testAccAppsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

data "fly_apps" "test" {
  depends_on = [fly_app.test]
}
`, name, testAccOrg())
}
