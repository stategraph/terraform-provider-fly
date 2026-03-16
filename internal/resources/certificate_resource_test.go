package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccCertificateResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")
	hostname := provider.RandName("tf-test") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig(appName, hostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_certificate.test", "id"),
					resource.TestCheckResourceAttr("fly_certificate.test", "hostname", hostname),
					resource.TestCheckResourceAttrSet("fly_certificate.test", "check_status"),
				),
			},
			{
				Config:   testAccCertificateConfig(appName, hostname),
				PlanOnly: true,
			},
		},
	})
}

func TestAccCertificateResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")
	hostname := provider.RandName("tf-test") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig(appName, hostname),
			},
			{
				ResourceName:      "fly_certificate.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_certificate.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.Attributes["hostname"], nil
				},
			},
		},
	})
}

func testAccCertificateConfig(appName, hostname string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_certificate" "test" {
  app      = fly_app.test.name
  hostname = %q
}
`, appName, testAccOrg(), hostname)
}
