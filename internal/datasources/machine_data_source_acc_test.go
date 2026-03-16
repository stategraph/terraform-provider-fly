package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMachineDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fly_machine.test", "region", "iad"),
					resource.TestCheckResourceAttr("data.fly_machine.test", "image", "registry.fly.io/flyctl-utils:latest"),
					resource.TestCheckResourceAttrSet("data.fly_machine.test", "state"),
				),
			},
		},
	})
}

func testAccMachineDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "registry.fly.io/flyctl-utils:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }
}

data "fly_machine" "test" {
  app = fly_app.test.name
  id  = fly_machine.test.id
}
`, name, testAccOrg())
}
