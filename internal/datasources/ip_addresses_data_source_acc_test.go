package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccIPAddressesDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressesDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_ip_addresses.test", "ip_addresses.#", 1),
				),
			},
		},
	})
}

func testAccIPAddressesDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_ip_address" "test" {
  app  = fly_app.test.name
  type = "v6"
}

data "fly_ip_addresses" "test" {
  app        = fly_app.test.name
  depends_on = [fly_ip_address.test]
}
`, name, testAccOrg())
}
