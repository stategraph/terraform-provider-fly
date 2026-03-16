package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccPostgresCluster_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresClusterConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_postgres_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("fly_postgres_cluster.test", "name"),
					resource.TestCheckResourceAttrSet("fly_postgres_cluster.test", "status"),
				),
			},
			{
				Config:   testAccPostgresClusterConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccPostgresClusterConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_postgres_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}
`, name, testAccOrg())
}
