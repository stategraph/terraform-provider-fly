package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMPGClustersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "fly_mpg_clusters" "test" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.fly_mpg_clusters.test", "clusters.#"),
			},
		},
	})
}
