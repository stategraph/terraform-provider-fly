package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccPostgresAttachment_basic(t *testing.T) {
	pgName := provider.RandName("tf-test")
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresAttachmentConfig(pgName, appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_postgres_attachment.test", "id"),
					resource.TestCheckResourceAttrSet("fly_postgres_attachment.test", "postgres_app"),
					resource.TestCheckResourceAttrSet("fly_postgres_attachment.test", "app"),
				),
			},
			{
				Config:   testAccPostgresAttachmentConfig(pgName, appName),
				PlanOnly: true,
			},
		},
	})
}

func testAccPostgresAttachmentConfig(pgName, appName string) string {
	return fmt.Sprintf(`
resource "fly_postgres_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}

resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_postgres_attachment" "test" {
  postgres_app = fly_postgres_cluster.test.name
  app          = fly_app.test.name
}
`, pgName, testAccOrg(), appName, testAccOrg())
}
