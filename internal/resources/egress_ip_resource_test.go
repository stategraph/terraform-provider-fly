package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccEgressIPResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEgressIPConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_egress_ip.test", "id"),
					resource.TestCheckResourceAttrSet("fly_egress_ip.test", "address"),
				),
			},
			{
				Config:   testAccEgressIPConfig(appName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccEgressIPResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEgressIPConfig(appName),
			},
			{
				ResourceName:      "fly_egress_ip.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_egress_ip.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccEgressIPConfig(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_egress_ip" "test" {
  app = fly_app.test.name
}
`, appName, testAccOrg())
}
