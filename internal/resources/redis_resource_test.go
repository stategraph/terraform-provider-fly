package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccRedisResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRedisConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_redis.test", "id"),
					resource.TestCheckResourceAttrSet("fly_redis.test", "name"),
					resource.TestCheckResourceAttrSet("fly_redis.test", "status"),
				),
			},
			{
				Config:   testAccRedisConfig(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccRedisConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_redis" "test" {
  name   = %q
  org    = %q
  region = "iad"
}
`, name, testAccOrg())
}
