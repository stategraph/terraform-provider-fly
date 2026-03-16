package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccExtensionResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")
	extName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig(appName, extName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_ext_sentry.test", "id"),
					resource.TestCheckResourceAttrSet("fly_ext_sentry.test", "name"),
					resource.TestCheckResourceAttrSet("fly_ext_sentry.test", "status"),
				),
			},
			{
				Config:   testAccExtensionConfig(appName, extName),
				PlanOnly: true,
			},
		},
	})
}

func testAccExtensionConfig(appName, extName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_ext_sentry" "test" {
  name = %q
  app  = fly_app.test.name
}
`, appName, testAccOrg(), extName)
}
