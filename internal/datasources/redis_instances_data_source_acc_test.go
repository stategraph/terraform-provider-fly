package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedisInstancesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "fly_redis_instances" "test" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.fly_redis_instances.test", "instances.#"),
			},
		},
	})
}
