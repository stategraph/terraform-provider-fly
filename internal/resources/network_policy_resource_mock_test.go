package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apimodels"
)

func TestNetworkPolicyResource_lifecycle(t *testing.T) {
	policy := apimodels.NetworkPolicy{
		ID:   "pol-abc123",
		Name: "allow-egress-http",
		Selector: apimodels.PolicySelector{
			All: true,
		},
		Rules: []apimodels.PolicyRule{
			{
				Action:    "allow",
				Direction: "egress",
				Ports: []apimodels.PolicyPort{
					{Protocol: "tcp", Port: 80},
					{Protocol: "tcp", Port: 443},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-np", Name: "np-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/np-test-app" && !strings.Contains(r.URL.Path, "/network_policies"):
			json.NewEncoder(w).Encode(apimodels.App{ID: "app-np", Name: "np-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/np-test-app":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/apps/np-test-app/network_policies":
			json.NewEncoder(w).Encode(policy)
		case r.Method == "GET" && r.URL.Path == "/apps/np-test-app/network_policies/pol-abc123":
			json.NewEncoder(w).Encode(policy)
		case r.Method == "DELETE" && r.URL.Path == "/apps/np-test-app/network_policies/pol-abc123":
			w.WriteHeader(http.StatusOK)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + server.URL + `"
}

resource "fly_app" "test" {
  name     = "np-test-app"
  org_slug = "personal"
}

resource "fly_network_policy" "test" {
  app  = fly_app.test.name
  name = "allow-egress-http"

  selector {
    all = true
  }

  rule {
    action    = "allow"
    direction = "egress"

    port {
      protocol = "tcp"
      port     = 80
    }

    port {
      protocol = "tcp"
      port     = 443
    }
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_network_policy.test", "id", "pol-abc123"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "name", "allow-egress-http"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "selector.all", "true"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.#", "1"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.action", "allow"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.direction", "egress"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.#", "2"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("fly_network_policy.test", "rule.0.port.0.port", "80"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}
