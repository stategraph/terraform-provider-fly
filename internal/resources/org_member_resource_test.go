package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccOrgMemberResource_basic(t *testing.T) {
	email := os.Getenv("FLY_TEST_INVITE_EMAIL")
	if email == "" {
		t.Skip("FLY_TEST_INVITE_EMAIL must be set for org_member acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccOrgMemberConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_org_member.test", "id"),
					resource.TestCheckResourceAttr("fly_org_member.test", "email", email),
				),
			},
			{
				Config:   testAccOrgMemberConfig(email),
				PlanOnly: true,
			},
		},
	})
}

func testAccOrgMemberConfig(email string) string {
	return fmt.Sprintf(`
resource "fly_org_member" "test" {
  org   = %q
  email = %q
}
`, testAccOrg(), email)
}
