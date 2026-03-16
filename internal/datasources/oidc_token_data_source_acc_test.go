package datasources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func TestAccOIDCTokenDataSource_basic(t *testing.T) {
	// OIDC tokens are only available inside Fly.io machines.
	if os.Getenv("FLY_TEST_OIDC") == "" {
		t.Skip("FLY_TEST_OIDC must be set for OIDC token acceptance tests (requires Fly.io machine environment)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
data "fly_oidc_token" "test" {
  aud = "https://example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fly_oidc_token.test", "token"),
				),
			},
		},
	})
}
