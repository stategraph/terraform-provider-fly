package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMPGAttachmentResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMPGAttachmentConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_mpg_attachment.test", "id"),
					resource.TestCheckResourceAttrSet("fly_mpg_attachment.test", "cluster_id"),
					resource.TestCheckResourceAttr("fly_mpg_attachment.test", "app", name+"-app"),
				),
			},
			{
				Config:   testAccMPGAttachmentConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMPGAttachmentConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_mpg_cluster" "test" {
  name   = %q
  org    = %q
  region = "iad"
}

resource "fly_app" "test" {
  name     = "%s-app"
  org_slug = %q
}

resource "fly_mpg_attachment" "test" {
  cluster_id = fly_mpg_cluster.test.id
  app        = fly_app.test.name
}
`, name, testAccOrg(), name, testAccOrg())
}
