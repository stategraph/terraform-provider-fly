package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMPGUserResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMPGUserConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_mpg_user.test", "id"),
					resource.TestCheckResourceAttr("fly_mpg_user.test", "username", name+"-user"),
					resource.TestCheckResourceAttrSet("fly_mpg_user.test", "cluster_id"),
				),
			},
			{
				Config:   testAccMPGUserConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMPGUserConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_mpg_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}

resource "fly_mpg_user" "test" {
  cluster_id = fly_mpg_cluster.test.id
  username   = "%s-user"
}
`, name, testAccOrg(), name)
}
