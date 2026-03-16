package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccAppDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAppDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_app.test", "name", name),
					resource.TestCheckResourceAttr("data.fly_app.test", "org_slug", testAccOrg()),
					resource.TestCheckResourceAttrSet("data.fly_app.test", "status"),
				),
			},
		},
	})
}

func testAccAppDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

data "fly_app" "test" {
  name = fly_app.test.name
}
`, name, testAccOrg())
}
