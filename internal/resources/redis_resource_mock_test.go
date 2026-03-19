package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRedisResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"redis create --name test-redis --org personal --region iad": {
			Stdout: "Created Redis database test-redis\n",
		},
		"redis status test-redis --json": {
			Stdout: `{"id":"redis-123","name":"test-redis","status":"running","region":"iad","plan":"free","primary_url":"redis://test-redis.internal:6379","replica_regions":["ord"],"enable_eviction":false}`,
		},
		"redis destroy test-redis --yes": {
			Stdout: "Destroyed test-redis\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_redis" "test" {
  name   = "test-redis"
  org    = "personal"
  region = "iad"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_redis.test", "id", "redis-123"),
					resource.TestCheckResourceAttr("fly_redis.test", "name", "test-redis"),
					resource.TestCheckResourceAttr("fly_redis.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_redis.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_redis.test", "plan", "free"),
					resource.TestCheckResourceAttr("fly_redis.test", "primary_url", "redis://test-redis.internal:6379"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
