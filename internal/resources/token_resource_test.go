package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccTokenResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")
	tokenName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTokenConfig(appName, tokenName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_token.test", "id"),
					resource.TestCheckResourceAttrSet("fly_token.test", "token"),
				),
			},
			{
				Config:   testAccTokenConfig(appName, tokenName),
				PlanOnly: true,
			},
		},
	})
}

func testAccTokenConfig(appName, tokenName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_token" "test" {
  type = "deploy"
  app  = fly_app.test.name
  name = %q
}
`, appName, testAccOrg(), tokenName)
}
