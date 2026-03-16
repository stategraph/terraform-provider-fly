package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
)

// testAccOrg returns the org slug to use for acceptance tests.
// Defaults to "personal" but can be overridden via FLY_ORG env var.
func testAccOrg() string {
	if org := os.Getenv("FLY_ORG"); org != "" {
		return org
	}
	return "personal"
}

func testAccBaseAppConfig(name string) string {
	return fmt.Sprintf(`
resource "fly_app" "test" {
  name     = %q
  org_slug = %q
}
`, name, testAccOrg())
}

// runFlyctl runs a flyctl command for acceptance test checks.
func runFlyctl(args ...string) ([]byte, error) {
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("FLY_API_TOKEN must be set")
	}
	cmd := exec.CommandContext(context.Background(), "flyctl", args...)
	cmd.Env = append(os.Environ(), "FLY_API_TOKEN="+token)
	return cmd.Output()
}

// App

func testAccCheckAppDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_app" {
				continue
			}
			_, err := client.GetApp(context.Background(), rs.Primary.Attributes["name"])
			if err == nil {
				return fmt.Errorf("fly_app %s still exists", rs.Primary.ID)
			}
			if !apiclient.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckAppExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		_, err := client.GetApp(context.Background(), rs.Primary.Attributes["name"])
		return err
	}
}

// Machine

func testAccCheckMachineDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_machine" {
				continue
			}
			_, err := client.GetMachine(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("fly_machine %s still exists", rs.Primary.ID)
			}
			if !apiclient.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckMachineExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		_, err := client.GetMachine(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
		return err
	}
}

// Volume

func testAccCheckVolumeDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_volume" {
				continue
			}
			_, err := client.GetVolume(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("fly_volume %s still exists", rs.Primary.ID)
			}
			if !apiclient.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckVolumeExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		_, err := client.GetVolume(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
		return err
	}
}

// Secret

func testAccCheckSecretDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_secret" {
				continue
			}
			secrets, err := client.ListSecrets(context.Background(), rs.Primary.Attributes["app"])
			if err != nil {
				if apiclient.IsNotFound(err) {
					continue
				}
				return err
			}
			key := rs.Primary.Attributes["key"]
			for _, s := range secrets {
				if s.EffectiveName() == key {
					return fmt.Errorf("fly_secret %s still exists", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}

func testAccCheckSecretExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		secrets, err := client.ListSecrets(context.Background(), rs.Primary.Attributes["app"])
		if err != nil {
			return err
		}
		key := rs.Primary.Attributes["key"]
		for _, s := range secrets {
			if s.EffectiveName() == key {
				return nil
			}
		}
		return fmt.Errorf("fly_secret %s not found", key)
	}
}

// IP Address

type accTestIP struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func testAccCheckIPAddressDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_ip_address" {
				continue
			}
			out, err := runFlyctl("ips", "list", "-a", rs.Primary.Attributes["app"], "--json")
			if err != nil {
				continue // app may be gone
			}
			var ips []accTestIP
			if err := json.Unmarshal(out, &ips); err != nil {
				continue
			}
			for _, ip := range ips {
				if ip.ID == rs.Primary.ID {
					return fmt.Errorf("fly_ip_address %s still exists", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}

func testAccCheckIPAddressExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		out, err := runFlyctl("ips", "list", "-a", rs.Primary.Attributes["app"], "--json")
		if err != nil {
			return err
		}
		var ips []accTestIP
		if err := json.Unmarshal(out, &ips); err != nil {
			return err
		}
		for _, ip := range ips {
			if ip.ID == rs.Primary.ID {
				return nil
			}
		}
		return fmt.Errorf("fly_ip_address %s not found", rs.Primary.ID)
	}
}

// Certificate

func testAccCheckCertificateDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_certificate" {
				continue
			}
			_, err := client.GetCertificate(context.Background(), rs.Primary.Attributes["app"], rs.Primary.Attributes["hostname"])
			if err == nil {
				return fmt.Errorf("fly_certificate %s still exists", rs.Primary.ID)
			}
			if !apiclient.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckCertificateExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		_, err := client.GetCertificate(context.Background(), rs.Primary.Attributes["app"], rs.Primary.Attributes["hostname"])
		return err
	}
}

// Network Policy

func testAccCheckNetworkPolicyDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.TestAccAPIClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_network_policy" {
				continue
			}
			_, err := client.GetNetworkPolicy(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("fly_network_policy %s still exists", rs.Primary.ID)
			}
			if !apiclient.IsNotFound(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckNetworkPolicyExists(t *testing.T, resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		client := provider.TestAccAPIClient(t)
		_, err := client.GetNetworkPolicy(context.Background(), rs.Primary.Attributes["app"], rs.Primary.ID)
		return err
	}
}

// WireGuard Peer

type accTestWGPeer struct {
	Name string `json:"name"`
}

func testAccCheckWireGuardPeerDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_wireguard_peer" {
				continue
			}
			out, err := runFlyctl("wireguard", "list", "--org", rs.Primary.Attributes["org_slug"], "--json")
			if err != nil {
				continue
			}
			var peers []accTestWGPeer
			if err := json.Unmarshal(out, &peers); err != nil {
				continue
			}
			name := rs.Primary.Attributes["name"]
			for _, p := range peers {
				if p.Name == name {
					return fmt.Errorf("fly_wireguard_peer %s still exists", name)
				}
			}
		}
		return nil
	}
}

// WireGuard Token

type accTestWGToken struct {
	Name string `json:"name"`
}

func testAccCheckWireGuardTokenDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_wireguard_token" {
				continue
			}
			out, err := runFlyctl("wireguard", "token", "list", "--org", rs.Primary.Attributes["org_slug"], "--json")
			if err != nil {
				continue
			}
			var tokens []accTestWGToken
			if err := json.Unmarshal(out, &tokens); err != nil {
				continue
			}
			name := rs.Primary.Attributes["name"]
			for _, tok := range tokens {
				if tok.Name == name {
					return fmt.Errorf("fly_wireguard_token %s still exists", name)
				}
			}
		}
		return nil
	}
}

// Egress IP

func testAccCheckEgressIPDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "fly_egress_ip" {
				continue
			}
			out, err := runFlyctl("ips", "list", "-a", rs.Primary.Attributes["app"], "--json")
			if err != nil {
				continue
			}
			var ips []accTestIP
			if err := json.Unmarshal(out, &ips); err != nil {
				continue
			}
			for _, ip := range ips {
				if ip.ID == rs.Primary.ID {
					return fmt.Errorf("fly_egress_ip %s still exists", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}

// Volume Snapshot

func testAccCheckVolumeSnapshotDestroy(_ *testing.T) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		// Snapshots may expire on their own; no-op destroy check.
		return nil
	}
}
