package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMPGDatabaseResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMPGDatabaseConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_mpg_database.test", "id"),
					resource.TestCheckResourceAttr("fly_mpg_database.test", "name", name+"-db"),
					resource.TestCheckResourceAttrSet("fly_mpg_database.test", "cluster_id"),
				),
			},
			{
				Config:   testAccMPGDatabaseConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMPGDatabaseConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_mpg_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}

resource "fly_mpg_database" "test" {
  cluster_id = fly_mpg_cluster.test.id
  name       = "%s-db"
}
`, name, testAccOrg(), name)
}
