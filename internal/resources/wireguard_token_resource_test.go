package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccWireGuardTokenResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWireGuardTokenConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_wireguard_token.test", "id"),
					resource.TestCheckResourceAttr("fly_wireguard_token.test", "name", name),
					resource.TestCheckResourceAttrSet("fly_wireguard_token.test", "token"),
				),
			},
			{
				Config:   testAccWireGuardTokenConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func TestAccWireGuardTokenResource_import(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWireGuardTokenConfig(name),
			},
			{
				ResourceName:            "fly_wireguard_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_wireguard_token.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["org_slug"] + "/" + rs.Primary.Attributes["name"], nil
				},
			},
		},
	})
}

func testAccWireGuardTokenConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_wireguard_token" "test" {
  org_slug = %q
  name     = %q
}
`, testAccOrg(), name)
}
