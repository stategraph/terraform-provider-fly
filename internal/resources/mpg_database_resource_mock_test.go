package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMPGDatabaseResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"mpg databases create --cluster-id mpg-123 --name mydb --json": {
			Stdout: `{"name":"mydb"}`,
		},
		"mpg databases list --cluster-id mpg-123 --json": {
			Stdout: `[{"name":"mydb"}]`,
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_mpg_database" "test" {
  cluster_id = "mpg-123"
  name       = "mydb"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_mpg_database.test", "id", "mydb"),
					resource.TestCheckResourceAttr("fly_mpg_database.test", "name", "mydb"),
					resource.TestCheckResourceAttr("fly_mpg_database.test", "cluster_id", "mpg-123"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
