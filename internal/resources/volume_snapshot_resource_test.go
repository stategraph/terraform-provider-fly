package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccVolumeSnapshotResource_basic(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeSnapshotConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_volume_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("fly_volume_snapshot.test", "status"),
					resource.TestCheckResourceAttrSet("fly_volume_snapshot.test", "created_at"),
				),
			},
			{
				Config:   testAccVolumeSnapshotConfig(appName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVolumeSnapshotResource_import(t *testing.T) {
	appName := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeSnapshotConfig(appName),
			},
			{
				ResourceName:            "fly_volume_snapshot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"size", "digest"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_volume_snapshot.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["app"] + "/" + rs.Primary.Attributes["volume_id"] + "/" + rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccVolumeSnapshotConfig(appName string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}

resource "fly_volume" "test" {
  app     = fly_app.test.name
  name    = "testdata"
  region  = "iad"
  size_gb = 1
}

resource "fly_volume_snapshot" "test" {
  app       = fly_app.test.name
  volume_id = fly_volume.test.id
}
`, appName, testAccOrg())
}
