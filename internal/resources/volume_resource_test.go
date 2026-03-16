package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccVolumeResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(appName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists(t, "fly_volume.test"),
					resource.TestCheckResourceAttrSet("fly_volume.test", "id"),
					resource.TestCheckResourceAttr("fly_volume.test", "name", "testdata"),
					resource.TestCheckResourceAttr("fly_volume.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "1"),
					resource.TestCheckResourceAttrSet("fly_volume.test", "state"),
				),
			},
			{
				Config:   testAccVolumeConfig(appName, 1),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVolumeResource_extend(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(appName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists(t, "fly_volume.test"),
					resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "1"),
				),
			},
			{
				Config: testAccVolumeConfig(appName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_volume.test", "size_gb", "2"),
				),
			},
		},
	})
}

func TestAccVolumeResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(appName, 1),
			},
			{
				ResourceName:            "fly_volume.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"snapshot_id", "source_volume_id", "require_unique_zone"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_volume.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccVolumeResource_disappears(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig(appName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists(t, "fly_volume.test"),
					func(s *terraform.State) error {
						client := provider.TestAccAPIClient(t)
						rs := s.RootModule().Resources["fly_volume.test"]
						return client.DeleteVolume(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVolumeConfig(appName string, sizeGB int) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_volume" "test" {
  app     = fly_app.test.name
  name    = "testdata"
  region  = "iad"
  size_gb = %d
}
`, appName, testAccOrg(), sizeGB)
}
