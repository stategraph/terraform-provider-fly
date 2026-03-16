package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWireGuardPeerResource_lifecycle(t *testing.T) {
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"wireguard create personal iad test-peer --pubkey test-pubkey --json": {
			Stdout: `{"name":"test-peer","region":"iad","pubkey":"test-pubkey","peerip":"fdaa::1","endpointip":"1.2.3.4","gatewayip":""}`,
		},
		"wireguard list --org personal --json": {
			Stdout: `[{"id":"wg-1","name":"test-peer","region":"iad","pubkey":"test-pubkey","peerip":"fdaa::1","endpointip":"1.2.3.4","gatewayip":""}]`,
		},
		"wireguard remove personal test-peer --yes": {
			Stdout: "Removed peer test-peer\n",
		},
	})

	config := providerConfigWithFlyctl("http://localhost:1", flyctlPath) + `
resource "fly_wireguard_peer" "test" {
  org_slug   = "personal"
  region     = "iad"
  name       = "test-peer"
  public_key = "test-pubkey"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_wireguard_peer.test", "id", "personal/test-peer"),
					resource.TestCheckResourceAttr("fly_wireguard_peer.test", "peer_ip", "fdaa::1"),
					resource.TestCheckResourceAttr("fly_wireguard_peer.test", "endpoint_ip", "1.2.3.4"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
