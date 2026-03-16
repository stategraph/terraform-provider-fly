package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestEgressIPResource_lifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-eg", Name: "egress-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/egress-test-app":
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-eg", Name: "egress-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/egress-test-app":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"ips allocate-egress -a egress-test-app --yes": {
			Stdout: "Allocated egress IP\n",
		},
		"ips list -a egress-test-app --json": {
			Stdout: `[{"id":"eip-1","address":"5.6.7.8","version":"v4","region":"iad","city":"Ashburn"}]`,
		},
		"ips release-egress 5.6.7.8 -a egress-test-app": {
			Stdout: "Released 5.6.7.8\n",
		},
	})

	config := providerConfigWithFlyctl(server.URL, flyctlPath) + `
resource "fly_app" "test" {
  name     = "egress-test-app"
  org_slug = "personal"
}

resource "fly_egress_ip" "test" {
  app = fly_app.test.name
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_egress_ip.test", "id", "eip-1"),
					resource.TestCheckResourceAttr("fly_egress_ip.test", "address", "5.6.7.8"),
					resource.TestCheckResourceAttr("fly_egress_ip.test", "version", "v4"),
					resource.TestCheckResourceAttr("fly_egress_ip.test", "region", "iad"),
					resource.TestCheckResourceAttr("fly_egress_ip.test", "city", "Ashburn"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
