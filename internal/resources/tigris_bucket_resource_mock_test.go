package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTigrisBucketResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"storage create --name test-bucket --org personal": {
			Stdout: "Created Tigris bucket test-bucket\n",
		},
		"storage status test-bucket --json": {
			Stdout: `{"id":"bucket-123","name":"test-bucket","status":"ready","public":false}`,
		},
		"storage destroy test-bucket --yes": {
			Stdout: "Destroyed test-bucket\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_tigris_bucket" "test" {
  name = "test-bucket"
  org  = "personal"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_tigris_bucket.test", "id", "bucket-123"),
					resource.TestCheckResourceAttr("fly_tigris_bucket.test", "name", "test-bucket"),
					resource.TestCheckResourceAttr("fly_tigris_bucket.test", "status", "ready"),
					resource.TestCheckResourceAttr("fly_tigris_bucket.test", "public", "false"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
