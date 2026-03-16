package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccTigrisBucketResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTigrisBucketConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_tigris_bucket.test", "id"),
					resource.TestCheckResourceAttrSet("fly_tigris_bucket.test", "name"),
					resource.TestCheckResourceAttrSet("fly_tigris_bucket.test", "status"),
				),
			},
			{
				Config:   testAccTigrisBucketConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccTigrisBucketConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_tigris_bucket" "test" {
  name = %q
  org  = %q
}
`, name, testAccOrg())
}
