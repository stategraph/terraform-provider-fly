package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTokensDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "fly_tokens" "test" {
  org = %q
}`, testAccOrg()),
				Check: resource.TestCheckResourceAttrSet("data.fly_tokens.test", "tokens.#"),
			},
		},
	})
}
