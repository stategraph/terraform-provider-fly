package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMPGClusterResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMPGClusterConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_mpg_cluster.test", "id"),
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "name", name),
					resource.TestCheckResourceAttrSet("fly_mpg_cluster.test", "status"),
				),
			},
			{
				Config:   testAccMPGClusterConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMPGClusterConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_mpg_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}
`, name, testAccOrg())
}
