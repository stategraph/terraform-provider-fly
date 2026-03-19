package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestExtMySQLResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"ext mysql create --name test-mysql --org personal --region iad": {
			Stdout: "Created ext mysql test-mysql\n",
		},
		"ext mysql status test-mysql --json": {
			Stdout: `{"id":"ext-1","name":"test-mysql","status":"running"}`,
		},
		"ext mysql destroy test-mysql --yes": {
			Stdout: "Destroyed test-mysql\n",
		},
	})

	mysqlConfig := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_ext_mysql" "test" {
  name   = "test-mysql"
  org    = "personal"
  region = "iad"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: mysqlConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_ext_mysql.test", "id", "ext-1"),
					resource.TestCheckResourceAttr("fly_ext_mysql.test", "name", "test-mysql"),
					resource.TestCheckResourceAttr("fly_ext_mysql.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_ext_mysql.test", "org", "personal"),
					resource.TestCheckResourceAttr("fly_ext_mysql.test", "region", "iad"),
				),
			},
			{
				Config:   mysqlConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestExtSentryResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"ext sentry create --name test-sentry --app my-app": {
			Stdout: "Created ext sentry test-sentry\n",
		},
		"ext sentry status test-sentry --json": {
			Stdout: `{"id":"ext-2","name":"test-sentry","status":"running"}`,
		},
		"ext sentry destroy test-sentry --yes": {
			Stdout: "Destroyed test-sentry\n",
		},
	})

	sentryConfig := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_ext_sentry" "test" {
  name = "test-sentry"
  app  = "my-app"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: sentryConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_ext_sentry.test", "id", "ext-2"),
					resource.TestCheckResourceAttr("fly_ext_sentry.test", "name", "test-sentry"),
					resource.TestCheckResourceAttr("fly_ext_sentry.test", "status", "running"),
					resource.TestCheckResourceAttr("fly_ext_sentry.test", "app", "my-app"),
				),
			},
			{
				Config:   sentryConfig,
				PlanOnly: true,
			},
		},
	})
}
