package resources_test

import (
	"context"
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

