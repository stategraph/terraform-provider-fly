package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPostgresClusterResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"postgres create --name test-pg --org personal --region iad": {
			Stdout: "Created Postgres cluster test-pg\n",
		},
		"postgres list --json": {
			Stdout: `[{"id":"pg-123","name":"test-pg","status":"running","region":"iad","org":"personal","cluster_size":1,"volume_size":10,"vm_size":"shared-cpu-1x","enable_backups":false}]`,
		},
		"postgres destroy test-pg --yes": {
			Stdout: "Destroyed test-pg\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_postgres_cluster" "test" {
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
					resource.TestCheckResourceAttr("fly_postgres_cluster.test", "id", "pg-123"),
					resource.TestCheckResourceAttr("fly_postgres_cluster.test", "name", "test-pg"),
					resource.TestCheckResourceAttr("fly_postgres_cluster.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_postgres_cluster.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_postgres_cluster.test", "org", "personal"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
