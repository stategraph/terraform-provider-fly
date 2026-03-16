package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWireGuardTokenResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"wireguard token create --org personal --name test-token --json": {
			Stdout: `{"id":"wgt-1","name":"test-token","token":"secret-token-value"}`,
		},
		"wireguard token list --org personal --json": {
			Stdout: `[{"id":"wgt-1","name":"test-token"}]`,
		},
		"wireguard token delete --org personal --name test-token --yes": {
			Stdout: "Deleted token test-token\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_wireguard_token" "test" {
  org_slug = "personal"
  name     = "test-token"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_wireguard_token.test", "id", "personal/test-token"),
					resource.TestCheckResourceAttr("fly_wireguard_token.test", "name", "test-token"),
					resource.TestCheckResourceAttr("fly_wireguard_token.test", "token", "secret-token-value"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
