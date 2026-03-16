package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMPGAttachmentResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"mpg attach --cluster-id mpg-123 --app myapp --json": {
			Stdout: `{"connection_uri":"postgres://user:pass@host:5432/db","variable_name":"DATABASE_URL"}`,
		},
		"mpg detach --cluster-id mpg-123 --app myapp --yes": {
			Stdout: "Detached mpg-123 from myapp\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_mpg_attachment" "test" {
  cluster_id = "mpg-123"
  app        = "myapp"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "id", "mpg-123/myapp"),
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "cluster_id", "mpg-123"),
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "app", "myapp"),
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "variable_name", "DATABASE_URL"),
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "connection_uri", "postgres://user:pass@host:5432/db"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
