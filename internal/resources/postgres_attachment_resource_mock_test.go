package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPostgresAttachmentResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"postgres attach test-pg --app my-app --database-name my_app --json": {
			Stdout: `{"connection_uri":"postgres://user:pass@test-pg.internal:5432/my_app?sslmode=disable","variable_name":"DATABASE_URL","database_name":"my_app"}`,
		},
		"postgres detach test-pg --app my-app --yes": {
			Stdout: "Detached test-pg from my-app\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_postgres_attachment" "test" {
  postgres_app  = "test-pg"
  app           = "my-app"
  database_name = "my_app"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_postgres_attachment.test", "id", "test-pg/my-app"),
					resource.TestCheckResourceAttr("fly_postgres_attachment.test", "postgres_app", "test-pg"),
					resource.TestCheckResourceAttr("fly_postgres_attachment.test", "app", "my-app"),
					resource.TestCheckResourceAttr("fly_postgres_attachment.test", "variable_name", "DATABASE_URL"),
					resource.TestCheckResourceAttr("fly_postgres_attachment.test", "database_name", "my_app"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
