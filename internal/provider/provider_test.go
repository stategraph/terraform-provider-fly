package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProvider_schema(t *testing.T) {
	_, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("failed to create provider server: %v", err)
	}
}

func TestProvider_configure_missingToken(t *testing.T) {
	t.Setenv("FLY_API_TOKEN", "")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"fly": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
provider "fly" {}

data "fly_app" "test" {
  name = "nonexistent"
}
`,
				ExpectError: regexp.MustCompile("Missing API Token"),
			},
		},
	})
}
