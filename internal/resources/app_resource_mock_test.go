package resources_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/internal/provider"
)

func testProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fly": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

// mockAppJSON returns a JSON response matching the real Fly.io Machines API format.
// POST /apps returns only {"id":"...","created_at":...}
// GET /apps/<name> returns the full object with nested organization.
func mockAppCreateJSON() map[string]any {
	return map[string]any{
		"id":         "app-test-123",
		"created_at": 1700000000000,
	}
}

func mockAppGetJSON() map[string]any {
	return map[string]any{
		"id":     "app-test-123",
		"name":   "mock-app",
		"status": "deployed",
		"network": "default",
		"organization": map[string]any{
			"name": "Personal",
			"slug": "personal",
		},
	}
}

func TestAppResource_lifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(mockAppCreateJSON())
		case r.Method == "GET" && r.URL.Path == "/apps/mock-app":
			_ = json.NewEncoder(w).Encode(mockAppGetJSON())
		case r.Method == "DELETE" && r.URL.Path == "/apps/mock-app":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := testAppConfigWithURL(server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_app.test", "name", "mock-app"),
					resource.TestCheckResourceAttr("fly_app.test", "org_slug", "personal"),
					resource.TestCheckResourceAttr("fly_app.test", "id", "app-test-123"),
					resource.TestCheckResourceAttr("fly_app.test", "network", "default"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAppResource_disappears(t *testing.T) {
	getCallCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(mockAppCreateJSON())
		case r.Method == "GET" && r.URL.Path == "/apps/mock-app":
			getCallCount++
			if getCallCount > 1 {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(mockAppGetJSON())
		case r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAppConfigWithURL(server.URL),
				Check:  resource.TestCheckResourceAttr("fly_app.test", "name", "mock-app"),
			},
			{
				Config:             testAppConfigWithURL(server.URL),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAppResource_createError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "app name already taken"})
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAppConfigWithURL(server.URL),
				ExpectError: regexp.MustCompile("app name already taken"),
			},
		},
	})
}

func testAppConfigWithURL(apiURL string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "mock-app"
  org_slug = "personal"
}
`
}
