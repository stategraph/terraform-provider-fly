package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccSecretResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecretDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig(appName, "MY_SECRET", "secret-value-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(t, "fly_secret.test"),
					resource.TestCheckResourceAttr("fly_secret.test", "app", appName),
					resource.TestCheckResourceAttr("fly_secret.test", "key", "MY_SECRET"),
					resource.TestCheckResourceAttrSet("fly_secret.test", "digest"),
					resource.TestCheckResourceAttrSet("fly_secret.test", "id"),
				),
			},
			{
				Config:   testAccSecretConfig(appName, "MY_SECRET", "secret-value-1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSecretResource_update(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecretDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig(appName, "MY_SECRET", "value-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(t, "fly_secret.test"),
					resource.TestCheckResourceAttrSet("fly_secret.test", "digest"),
				),
			},
			{
				Config: testAccSecretConfig(appName, "MY_SECRET", "value-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(t, "fly_secret.test"),
					resource.TestCheckResourceAttrSet("fly_secret.test", "digest"),
				),
			},
		},
	})
}

func TestAccSecretResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecretDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig(appName, "MY_SECRET", "secret-value-1"),
			},
			{
				ResourceName:            "fly_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_secret.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/MY_SECRET", nil
				},
			},
		},
	})
}

func TestAccSecretResource_disappears(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecretDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig(appName, "MY_SECRET", "secret-value-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(t, "fly_secret.test"),
					func(s *terraform.State) error {
						client := provider.TestAccAPIClient(t)
						return client.DeleteSecret(context.Background(), appName, "MY_SECRET")
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSecretConfig(appName, key, value string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_secret" "test" {
  app   = fly_app.test.name
  key   = %q
  value = %q
}
`, appName, testAccOrg(), key, value)
}
