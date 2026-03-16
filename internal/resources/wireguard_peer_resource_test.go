package resources_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func testGenerateWireGuardKey() string {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func TestAccWireGuardPeerResource_basic(t *testing.T) {
	name := provider.RandName("tf-test")
	pubkey := testGenerateWireGuardKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWireGuardPeerConfig(name, pubkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fly_wireguard_peer.test", "id"),
					resource.TestCheckResourceAttrSet("fly_wireguard_peer.test", "peer_ip"),
					resource.TestCheckResourceAttrSet("fly_wireguard_peer.test", "endpoint_ip"),
				),
			},
			{
				Config:   testAccWireGuardPeerConfig(name, pubkey),
				PlanOnly: true,
			},
		},
	})
}

func TestAccWireGuardPeerResource_import(t *testing.T) {
	name := provider.RandName("tf-test")
	pubkey := testGenerateWireGuardKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWireGuardPeerConfig(name, pubkey),
			},
			{
				ResourceName:            "fly_wireguard_peer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["fly_wireguard_peer.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return testAccOrg() + "/" + rs.Primary.Attributes["name"], nil
				},
			},
		},
	})
}

func testAccWireGuardPeerConfig(name, pubkey string) string {
	return fmt.Sprintf(`
resource "fly_wireguard_peer" "test" {
  org_slug   = %q
  region     = "iad"
  name       = %q
  public_key = %q
}
`, testAccOrg(), name, pubkey)
}
