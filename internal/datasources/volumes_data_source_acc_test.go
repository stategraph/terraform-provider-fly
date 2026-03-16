package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccVolumesDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumesDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_volumes.test", "volumes.#", 1),
				),
			},
		},
	})
}

func testAccVolumesDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_volume" "test" {
  app     = fly_app.test.name
  name    = "testdata"
  region  = "iad"
  size_gb = 1
}

data "fly_volumes" "test" {
  app        = fly_app.test.name
  depends_on = [fly_volume.test]
}
`, name, testAccOrg())
}
