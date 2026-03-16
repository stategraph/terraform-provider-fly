package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTokenResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"tokens create deploy --json --org personal --name test-token": {
			Stdout: `{"id":"token-123","name":"test-token","type":"deploy","token":"fm2_supersecretvalue"}`,
		},
		"tokens list --json --org personal": {
			Stdout: `[{"id":"token-123","name":"test-token","type":"deploy"}]`,
		},
		"tokens revoke token-123 --yes": {
			Stdout: "Revoked token-123\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_token" "test" {
  type = "deploy"
  org  = "personal"
  name = "test-token"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_token.test", "id", "token-123"),
					resource.TestCheckResourceAttr("fly_token.test", "name", "test-token"),
					resource.TestCheckResourceAttr("fly_token.test", "token", "fm2_supersecretvalue"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
