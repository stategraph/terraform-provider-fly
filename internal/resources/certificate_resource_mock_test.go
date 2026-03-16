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

func TestCertificateResource_lifecycle(t *testing.T) {
	cert := apimodels.Certificate{
		ID:                    "cert-123",
		Hostname:              "example.com",
		CheckStatus:           "passing",
		DNSValidationHostname: "_acme.example.com",
		DNSValidationTarget:   "example.com.flydns.net",
		Source:                "fly",
		CertificateAuthority:  "lets_encrypt",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/apps":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-cert", Name: "cert-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "GET" && r.URL.Path == "/apps/cert-test-app" && !strings.Contains(r.URL.Path, "/certificates"):
			_ = json.NewEncoder(w).Encode(apimodels.App{ID: "app-cert", Name: "cert-test-app", Organization: apimodels.AppOrg{Slug: "personal"}, Network: "default", Status: "deployed"})
		case r.Method == "DELETE" && r.URL.Path == "/apps/cert-test-app":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "POST" && r.URL.Path == "/apps/cert-test-app/certificates":
			_ = json.NewEncoder(w).Encode(cert)
		case r.Method == "GET" && r.URL.Path == "/apps/cert-test-app/certificates/example.com":
			_ = json.NewEncoder(w).Encode(cert)
		case r.Method == "DELETE" && r.URL.Path == "/apps/cert-test-app/certificates/example.com":
			w.WriteHeader(http.StatusOK)

		default:
			t.Logf("Unhandled: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	config := testCertConfigWithURL(server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fly_certificate.test", "id", "cert-123"),
					resource.TestCheckResourceAttr("fly_certificate.test", "hostname", "example.com"),
					resource.TestCheckResourceAttr("fly_certificate.test", "check_status", "passing"),
					resource.TestCheckResourceAttr("fly_certificate.test", "source", "fly"),
					resource.TestCheckResourceAttr("fly_certificate.test", "dns_validation_hostname", "_acme.example.com"),
					resource.TestCheckResourceAttr("fly_certificate.test", "dns_validation_target", "example.com.flydns.net"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testCertConfigWithURL(apiURL string) string {
	return `
provider "fly" {
  api_token = "mock-token"
  api_url   = "` + apiURL + `"
}

resource "fly_app" "test" {
  name     = "cert-test-app"
  org_slug = "personal"
}

resource "fly_certificate" "test" {
  app      = fly_app.test.name
  hostname = "example.com"
}
`
}
