package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"orgs create test-org": {
			Stdout: "Created organization test-org\n",
		},
		"orgs show test-org --json": {
			Stdout: `{"id":"org-123","name":"test-org","slug":"test-org"}`,
		},
		"orgs delete test-org --yes": {
			Stdout: "Deleted test-org\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_org" "test" {
  name = "test-org"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_org.test", "id", "org-123"),
					resource.TestCheckResourceAttr("fly_org.test", "name", "test-org"),
					resource.TestCheckResourceAttr("fly_org.test", "slug", "test-org"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
