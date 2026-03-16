package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestIPAddressResource_lifecycle(t *testing.T) {
	// Mock REST server for the fly_app dependency.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-ip", Name: "ip-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/ip-test-app":
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-ip", Name: "ip-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/ip-test-app":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	// Mock flyctl for IP operations.
	flyctlPath := createMockFlyctl(t, map[string]flyctlMockResponse{
		"ips allocate-v4 -a ip-test-app --shared": {
			Stdout: "Allocated 66.241.124.1\n",
		},
		"ips list -a ip-test-app --json": {
			Stdout: `[{"id":"ip-alloc-123","address":"66.241.124.1","type":"shared_v4","region":"","created_at":"2024-01-01T00:00:00Z"}]`,
		},
		"ips release 66.241.124.1 -a ip-test-app": {
			Stdout: "Released 66.241.124.1\n",
		},
	})

	config := providerConfigWithFlyctl(server.URL, flyctlPath) + `
resource "fly_app" "test" {
  name     = "ip-test-app"
  org_slug = "personal"
}

resource "fly_ip_address" "test" {
  app  = fly_app.test.name
  type = "shared_v4"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_ip_address.test", "id", "ip-alloc-123"),
					resource.TestCheckResourceAttr("fly_ip_address.test", "address", "66.241.124.1"),
					resource.TestCheckResourceAttr("fly_ip_address.test", "type", "shared_v4"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
