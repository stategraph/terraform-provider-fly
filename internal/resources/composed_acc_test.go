package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccComposed_machineWithVolume(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComposedMachineWithVolumeConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_machine.test", "mount.0.volume"),
					resource.TestCheckResourceAttr("fly_machine.test", "mount.0.path", "/data"),
				),
			},
		},
	})
}

func TestAccComposed_appWithCertsAndIPs(t *testing.T) {
	appName := provider.RandName("tf-test")
	hostname := provider.RandName("tf-test") + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComposedAppWithCertsAndIPsConfig(appName, hostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_ip_address.test", "address"),
					resource.TestCheckResourceAttr("fly_certificate.test", "hostname", hostname),
				),
			},
		},
	})
}

func TestAccComposed_appWithSecretsAndMachine(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAppDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComposedAppWithSecretsAndMachineConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_secret.db", "digest"),
					resource.TestCheckResourceAttrSet("fly_secret.api", "digest"),
					resource.TestCheckResourceAttrSet("fly_machine.test", "id"),
				),
			},
		},
	})
}

func testAccComposedMachineWithVolumeConfig(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_volume" "test" {
  app     = fly_app.test.name
  name    = "data"
  region  = "iad"
  size_gb = 1
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

  mount {
    volume = fly_volume.test.id
    path   = "/data"
  }
}
`, appName, testAccOrg())
}

func testAccComposedAppWithCertsAndIPsConfig(appName, hostname string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_ip_address" "test" {
  app  = fly_app.test.name
  type = "v6"
}

resource "fly_certificate" "test" {
  app      = fly_app.test.name
  hostname = %q
}
`, appName, testAccOrg(), hostname)
}

func testAccComposedAppWithSecretsAndMachineConfig(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_secret" "db" {
  app   = fly_app.test.name
  key   = "DATABASE_URL"
  value = "postgres://test"
}

resource "fly_secret" "api" {
  app   = fly_app.test.name
  key   = "API_KEY"
  value = "test-key"
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

  depends_on = [fly_secret.db, fly_secret.api]
}
`, appName, testAccOrg())
}
