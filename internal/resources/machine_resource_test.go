package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccMachineResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMachineDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineConfig_basic(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMachineExists(t, "fly_machine.test"),
					resource.TestCheckResourceAttrSet("fly_machine.test", "id"),
					resource.TestCheckResourceAttr("fly_machine.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_machine.test", "image", "registry.fly.io/flyctl-utils:latest"),
					resource.TestCheckResourceAttrSet("fly_machine.test", "state"),
					resource.TestCheckResourceAttrSet("fly_machine.test", "private_ip"),
				),
			},
			{
				Config:   testAccMachineConfig_basic(appName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccMachineResource_withServices(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMachineDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineConfig_withServices(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMachineExists(t, "fly_machine.test"),
					resource.TestCheckResourceAttrSet("fly_machine.test", "id"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("fly_machine.test", "service.0.internal_port", "8080"),
				),
			},
		},
	})
}

func TestAccMachineResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMachineDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineConfig_basic(appName),
			},
			{
				ResourceName:            "fly_machine.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_launch", "desired_status"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_machine.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccMachineResource_update(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMachineDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineConfig_update(appName, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMachineExists(t, "fly_machine.test"),
					resource.TestCheckResourceAttr("fly_machine.test", "env.TEST", "true"),
				),
			},
			{
				Config: testAccMachineConfig_update(appName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMachineExists(t, "fly_machine.test"),
					resource.TestCheckResourceAttr("fly_machine.test", "env.TEST", "updated"),
				),
			},
		},
	})
}

func TestAccMachineResource_disappears(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMachineDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMachineConfig_basic(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMachineExists(t, "fly_machine.test"),
					func(s *terraform.State) error {
						client := provider.TestAccAPIClient(t)
						rs := s.RootModule().Resources["fly_machine.test"]
						_ = client.StopMachine(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
						_ = client.DeleteMachine(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
						return nil
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMachineConfig_basic(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "registry.fly.io/flyctl-utils:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  env = {
    TEST = "true"
  }
}
`, appName, testAccOrg())
}

func testAccMachineConfig_withServices(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "registry.fly.io/flyctl-utils:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  service {
    protocol      = "tcp"
    internal_port = 8080

    port {
      port     = 80
      handlers = ["http"]
    }

    port {
      port     = 443
      handlers = ["tls", "http"]
    }
  }
}
`, appName, testAccOrg())
}

func testAccMachineConfig_update(appName, envValue string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_machine" "test" {
  app    = fly_app.test.name
  region = "iad"
  image  = "registry.fly.io/flyctl-utils:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  env = {
    TEST = %q
  }
}
`, appName, testAccOrg(), envValue)
}
