package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgMemberResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"orgs invite personal user@example.com": {
			Stdout: "Invitation sent\n",
		},
		"orgs remove personal user@example.com --yes": {
			Stdout: "Removed user@example.com\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_org_member" "test" {
  org   = "personal"
  email = "user@example.com"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_org_member.test", "id", "personal/user@example.com"),
					resource.TestCheckResourceAttr("fly_org_member.test", "org", "personal"),
					resource.TestCheckResourceAttr("fly_org_member.test", "email", "user@example.com"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
