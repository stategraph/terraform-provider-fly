package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccVolumeDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_volume.test", "name", "testdata"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "size_gb", "1"),
					resource.TestCheckResourceAttr("data.fly_volume.test", "region", "iad"),
				),
			},
		},
	})
}

func testAccVolumeDataSourceConfig(name string) string {
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

data "fly_volume" "test" {
  app = fly_app.test.name
  id  = fly_volume.test.id
}
`, name, testAccOrg())
}
