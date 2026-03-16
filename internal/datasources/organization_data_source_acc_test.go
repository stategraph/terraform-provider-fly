package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fly_organization.test", "id"),
					resource.TestCheckResourceAttrSet("data.fly_organization.test", "name"),
					resource.TestCheckResourceAttr("data.fly_organization.test", "slug", testAccOrg()),
				),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig() string {
	return fmt.Sprintf(`
data "fly_organization" "test" {
  slug = %q
}
`, testAccOrg())
}
