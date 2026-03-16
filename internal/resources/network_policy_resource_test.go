package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccNetworkPolicyResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyConfig(appName, 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_network_policy.test", "id"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "name", "test-policy"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.action", "allow"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.direction", "egress"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.port", "80"),
				),
			},
			{
				Config:   testAccNetworkPolicyConfig(appName, 80),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNetworkPolicyResource_update(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyConfig(appName, 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.port", "80"),
				),
			},
			{
				Config: testAccNetworkPolicyConfig(appName, 443),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.port", "443"),
				),
			},
		},
	})
}

func TestAccNetworkPolicyResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPolicyConfig(appName, 80),
			},
			{
				ResourceName:      "fly_network_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_network_policy.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccNetworkPolicyConfig(appName string, port int) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_network_policy" "test" {
  app  = fly_app.test.name
  name = "test-policy"

  selector {
    all = true
  }

  rule {
    action    = "allow"
    direction = "egress"

    port {
      protocol = "tcp"
      port     = %d
    }
  }
}
`, appName, testAccOrg(), port)
}
