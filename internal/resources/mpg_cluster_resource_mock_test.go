package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMPGClusterResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"mpg create --name test-pg --org personal --region iad": {
			Stdout: "Created MPG cluster test-pg\n",
		},
		"mpg list --org personal --json": {
			Stdout: `[{"id":"mpg-123","name":"test-pg","status":"running","primary_region":"iad","region":"iad","plan":"starter","volume_size":10,"pg_major_version":16,"enable_postgis":false}]`,
		},
		"mpg destroy mpg-123 --yes": {
			Stdout: "Destroyed test-pg\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_mpg_cluster" "test" {
  name   = "test-pg"
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
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "id", "mpg-123"),
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "name", "test-pg"),
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "plan", "starter"),
					resource.TestCheckResourceAttr("fly_mpg_cluster.test", "pg_major_version", "16"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
