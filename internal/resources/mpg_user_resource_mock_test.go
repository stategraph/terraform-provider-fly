package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMPGUserResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"mpg users create --cluster-id mpg-123 --username testuser --json": {
			Stdout: `{"username":"testuser","role":"user"}`,
		},
		"mpg users list --cluster-id mpg-123 --json": {
			Stdout: `[{"username":"testuser","role":"user"}]`,
		},
		"mpg users delete --cluster-id mpg-123 --username testuser --yes": {
			Stdout: "Deleted testuser\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_mpg_user" "test" {
  cluster_id = "mpg-123"
  username   = "testuser"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_mpg_user.test", "id", "testuser"),
					resource.TestCheckResourceAttr("fly_mpg_user.test", "username", "testuser"),
					resource.TestCheckResourceAttr("fly_mpg_user.test", "cluster_id", "mpg-123"),
					resource.TestCheckResourceAttr("fly_mpg_user.test", "role", "user"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
