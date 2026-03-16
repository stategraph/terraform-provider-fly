package datasources_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fly": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("FLY_API_TOKEN"); v == "" {
		t.Fatal("FLY_API_TOKEN must be set for acceptance tests")
	}
}

func testAccOrg() string {
	if org := os.Getenv("FLY_ORG"); org != "" {
		return org
	}
	return "personal"
}

func testCheckAttrGreaterThan(name, key string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		count, _ := strconv.Atoi(rs.Primary.Attributes[key])
		if count < min {
			return fmt.Errorf("%s: expected %s >= %d, got %d", name, key, min, count)
		}
		return nil
	}
}
