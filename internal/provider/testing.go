package provider

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
)

func TestAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fly": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("FLY_API_TOKEN"); v == "" {
		t.Fatal("FLY_API_TOKEN must be set for acceptance tests")
	}
}

func RandName(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}

func TestAccAPIClient(t *testing.T) *apiclient.Client {
	t.Helper()
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		t.Fatal("FLY_API_TOKEN must be set for acceptance tests")
	}
	return apiclient.NewClient(token, "test")
}

func ConfigCompose(configs ...string) string {
	return strings.Join(configs, "\n")
}
