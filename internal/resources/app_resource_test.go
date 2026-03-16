package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccAppResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(t, "fly_app.test"),
					resource.TestCheckResourceAttr("fly_app.test", "name", appName),
					resource.TestCheckResourceAttr("fly_app.test", "org_slug", testAccOrg()),
					resource.TestCheckResourceAttrSet("fly_app.test", "id"),
					resource.TestCheckResourceAttrSet("fly_app.test", "status"),
				),
			},
			{
				Config:   testAccAppConfig(appName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAppResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig(appName),
			},
			{
				ResourceName:      "fly_app.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppResource_disappears(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(t, "fly_app.test"),
					func(s *terraform.State) error {
						client := provider.TestAccAPIClient(t)
						return client.DeleteApp(context.Background(), appName)
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAppConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}
`, name, testAccOrg())
}
