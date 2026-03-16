package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccCertificatesDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")
	hostname := provider.RandName("tf-test") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificatesDataSourceConfig(name, hostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_certificates.test", "certificates.#", 1),
				),
			},
		},
	})
}

func testAccCertificatesDataSourceConfig(name, hostname string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_certificate" "test" {
  app      = fly_app.test.name
  hostname = %q
}

data "fly_certificates" "test" {
  app        = fly_app.test.name
  depends_on = [fly_certificate.test]
}
`, name, testAccOrg(), hostname)
}
