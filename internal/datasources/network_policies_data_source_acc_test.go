package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccNetworkPoliciesDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkPoliciesDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_network_policies.test", "policies.#", 1),
				),
			},
		},
	})
}

func testAccNetworkPoliciesDataSourceConfig(name string) string {
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
      port     = 80
    }
  }
}

data "fly_network_policies" "test" {
  app        = fly_app.test.name
  depends_on = [fly_network_policy.test]
}
`, name, testAccOrg())
}
