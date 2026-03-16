package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccOrg_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_org.test", "id"),
					resource.TestCheckResourceAttrSet("fly_org.test", "name"),
					resource.TestCheckResourceAttrSet("fly_org.test", "slug"),
				),
			},
			{
				Config:   testAccOrgConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccOrgConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_org" "test" {
  name = %q
}
`, name)
}
