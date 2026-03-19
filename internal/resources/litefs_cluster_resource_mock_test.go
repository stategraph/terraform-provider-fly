package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLiteFSClusterResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"litefs-cloud clusters create --name test-litefs --org personal --region iad": {
			Stdout: "Created LiteFS cluster test-litefs\n",
		},
		"litefs-cloud clusters list --json": {
			Stdout: `[{"id":"lfs-123","name":"test-litefs","status":"running","region":"iad","org":"personal"}]`,
		},
		"litefs-cloud clusters destroy test-litefs --yes": {
			Stdout: "Destroyed test-litefs\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_litefs_cluster" "test" {
  name   = "test-litefs"
  org    = "personal"
  region = "iad"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_litefs_cluster.test", "id", "lfs-123"),
					resource.TestCheckResourceAttr("fly_litefs_cluster.test", "name", "test-litefs"),
					resource.TestCheckResourceAttr("fly_litefs_cluster.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_litefs_cluster.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_litefs_cluster.test", "org", "personal"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
