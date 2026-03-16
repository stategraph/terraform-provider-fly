package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMachinesDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMachinesDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_machines.test", "machines.#", 1),
				),
			},
		},
	})
}

func testAccMachinesDataSourceConfig(name string) string {
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

data "fly_machines" "test" {
  app        = fly_app.test.name
  depends_on = [fly_machine.test]
}
`, name, testAccOrg())
}
