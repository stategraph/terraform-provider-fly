package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccIPAddressResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfig(appName, "v6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_ip_address.test", "id"),
					resource.TestCheckResourceAttrSet("fly_ip_address.test", "address"),
					resource.TestCheckResourceAttr("fly_ip_address.test", "type", "v6"),
				),
			},
			{
				Config:   testAccIPAddressConfig(appName, "v6"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccIPAddressResource_shared_v4(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfig(appName, "shared_v4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_ip_address.test", "id"),
					resource.TestCheckResourceAttr("fly_ip_address.test", "type", "shared_v4"),
				),
			},
		},
	})
}

func TestAccIPAddressResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfig(appName, "v6"),
			},
			{
				ResourceName:      "fly_ip_address.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_ip_address.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccIPAddressResource_disappears(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfig(appName, "v6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fly_ip_address.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						_, err := runFlyctl("ips", "release", rs.Primary.Attributes["address"], "-a", rs.Primary.Attributes["app"], "--yes")
						return err
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIPAddressConfig(appName, ipType string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_ip_address" "test" {
  app  = fly_app.test.name
  type = %q
}
`, appName, testAccOrg(), ipType)
}
