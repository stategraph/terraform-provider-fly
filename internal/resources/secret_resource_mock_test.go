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

func TestSecretResource_lifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// App CRUD
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": "app-sec", "created_at": 1700000000000})
		case r.Method == "GET" && r.URL.Path == "/apps/secret-test-app" && !strings.Contains(r.URL.Path, "/secrets"):
			json.NewEncoder(w).Encode(map[string]any{
				"id": "app-sec", "name": "secret-test-app", "status": "deployed",
				"network": "default", "organization": map[string]any{"slug": "personal"},
			})
		case r.Method == "DELETE" && r.URL.Path == "/apps/secret-test-app":
			w.WriteHeader(http.StatusNoContent)

		// Set secret (POST /apps/{app}/secrets/{key})
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/apps/secret-test-app/secrets/"):
			json.NewEncoder(w).Encode(map[string]any{
				"name": "MY_SECRET", "value": "secret-value-1",
				"digest": "abc123hash", "version": 1,
			})

		// List secrets (GET /apps/{app}/secrets)
		case r.Method == "GET" && r.URL.Path == "/apps/secret-test-app/secrets":
			json.NewEncoder(w).Encode(map[string]any{
				"secrets": []apimodels.Secret{
					{Name: "MY_SECRET", Digest: "abc123hash", CreatedAt: "2024-01-01T00:00:00Z"},
				},
			})

		// Delete secret
		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/apps/secret-test-app/secrets/"):
			w.WriteHeader(http.StatusOK)

		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := testSecretConfigWithURL(server.URL, "secret-value-1")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_secret.test", "app", "secret-test-app"),
					resource.TestCheckResourceAttr("fly_secret.test", "key", "MY_SECRET"),
					resource.TestCheckResourceAttr("fly_secret.test", "id", "secret-test-app/MY_SECRET"),
					resource.TestCheckResourceAttrSet("fly_secret.test", "digest"),
					resource.TestCheckResourceAttr("fly_secret.test", "created_at", "2024-01-01T00:00:00Z"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testSecretConfigWithURL(apiURL, value string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "secret-test-app"
  org_slug = "personal"
}

resource "fly_secret" "test" {
  app   = fly_app.test.name
  key   = "MY_SECRET"
  value = "` + value + `"
}
`
}
