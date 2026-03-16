package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccVolumeSnapshotsDataSource_basic(t *testing.T) {
	name := provider.RandName("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeSnapshotsDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAttrGreaterThan("data.fly_volume_snapshots.test", "snapshots.#", 1),
				),
			},
		},
	})
}

func testAccVolumeSnapshotsDataSourceConfig(name string) string {
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

data "fly_volume_snapshots" "test" {
  app        = fly_app.test.name
  volume_id  = fly_volume.test.id
  depends_on = [fly_volume_snapshot.test]
}
`, name, testAccOrg())
}
