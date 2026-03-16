package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stategraph/terraform-provider-fly/pkg/apiclient"
)

func init() {
	resource.AddTestSweepers("fly_machine", &resource.Sweeper{
		Name: "fly_machine",
		F:    sweepMachines,
	})

	resource.AddTestSweepers("fly_volume", &resource.Sweeper{
		Name:         "fly_volume",
		Dependencies: []string{"fly_machine"},
		F:            sweepVolumes,
	})

	resource.AddTestSweepers("fly_ip_address", &resource.Sweeper{
		Name: "fly_ip_address",
		F:    sweepIPAddresses,
	})

	resource.AddTestSweepers("fly_certificate", &resource.Sweeper{
		Name: "fly_certificate",
		F:    sweepCertificates,
	})

	resource.AddTestSweepers("fly_network_policy", &resource.Sweeper{
		Name: "fly_network_policy",
		F:    sweepNetworkPolicies,
	})

	resource.AddTestSweepers("fly_wireguard_peer", &resource.Sweeper{
		Name: "fly_wireguard_peer",
		F:    sweepWireGuardPeers,
	})

	resource.AddTestSweepers("fly_wireguard_token", &resource.Sweeper{
		Name: "fly_wireguard_token",
		F:    sweepWireGuardTokens,
	})

	resource.AddTestSweepers("fly_app", &resource.Sweeper{
		Name:         "fly_app",
		Dependencies: []string{"fly_machine", "fly_volume", "fly_ip_address", "fly_certificate", "fly_network_policy"},
		F:            sweepApps,
	})

	resource.AddTestSweepers("fly_mpg_cluster", &resource.Sweeper{
		Name: "fly_mpg_cluster",
		F:    sweepMPGClusters,
	})

	resource.AddTestSweepers("fly_redis", &resource.Sweeper{
		Name: "fly_redis",
		F:    sweepRedis,
	})

	resource.AddTestSweepers("fly_tigris_bucket", &resource.Sweeper{
		Name: "fly_tigris_bucket",
		F:    sweepTigrisBuckets,
	})

	resource.AddTestSweepers("fly_postgres_cluster", &resource.Sweeper{
		Name: "fly_postgres_cluster",
		F:    sweepPostgresClusters,
	})

	resource.AddTestSweepers("fly_org", &resource.Sweeper{
		Name:         "fly_org",
		Dependencies: []string{"fly_app"},
		F:            sweepOrgs,
	})

	resource.AddTestSweepers("fly_litefs_cluster", &resource.Sweeper{
		Name: "fly_litefs_cluster",
		F:    sweepLiteFSClusters,
	})

	resource.AddTestSweepers("fly_ext_mysql", &resource.Sweeper{
		Name: "fly_ext_mysql",
		F:    sweepExtMySQL,
	})

	resource.AddTestSweepers("fly_ext_sentry", &resource.Sweeper{
		Name: "fly_ext_sentry",
		F:    sweepExtSentry,
	})

	resource.AddTestSweepers("fly_ext_kubernetes", &resource.Sweeper{
		Name: "fly_ext_kubernetes",
		F:    makeExtSweeper("kubernetes"),
	})

	resource.AddTestSweepers("fly_ext_arcjet", &resource.Sweeper{
		Name: "fly_ext_arcjet",
		F:    makeExtSweeper("arcjet"),
	})

	resource.AddTestSweepers("fly_ext_wafris", &resource.Sweeper{
		Name: "fly_ext_wafris",
		F:    makeExtSweeper("wafris"),
	})

	resource.AddTestSweepers("fly_ext_vector", &resource.Sweeper{
		Name: "fly_ext_vector",
		F:    makeExtSweeper("vector"),
	})
}

func sweepOrg() string {
	if org := os.Getenv("FLY_ORG"); org != "" {
		return org
	}
	return "personal"
}

func sweepClient() (*apiclient.Client, error) {
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("FLY_API_TOKEN must be set for sweeping")
	}
	return apiclient.NewClient(token, "test"), nil
}

// runFlyctl runs a flyctl command for sweep operations.
func runFlyctl(args ...string) ([]byte, error) {
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("FLY_API_TOKEN must be set for sweeping")
	}
	cmd := exec.CommandContext(context.Background(), "flyctl", args...)
	cmd.Env = append(os.Environ(), "FLY_API_TOKEN="+token)
	return cmd.Output()
}

func sweepApps(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for sweep: %w", err)
	}

	for _, app := range apps {
		if strings.HasPrefix(app.Name, "tf-test-") {
			fmt.Printf("Sweeping app: %s\n", app.Name)
			if err := client.DeleteApp(context.Background(), app.Name); err != nil {
				fmt.Printf("Error sweeping app %s: %v\n", app.Name, err)
			}
		}
	}

	return nil
}

func sweepMachines(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for machine sweep: %w", err)
	}

	for _, app := range apps {
		if !strings.HasPrefix(app.Name, "tf-test-") {
			continue
		}

		machines, err := client.ListMachines(context.Background(), app.Name)
		if err != nil {
			fmt.Printf("Error listing machines for app %s: %v\n", app.Name, err)
			continue
		}

		for _, machine := range machines {
			fmt.Printf("Sweeping machine %s in app %s\n", machine.ID, app.Name)
			_ = client.StopMachine(context.Background(), app.Name, machine.ID)
			if err := client.DeleteMachine(context.Background(), app.Name, machine.ID); err != nil {
				fmt.Printf("Error sweeping machine %s: %v\n", machine.ID, err)
			}
		}
	}

	return nil
}

func sweepVolumes(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for volume sweep: %w", err)
	}

	for _, app := range apps {
		if !strings.HasPrefix(app.Name, "tf-test-") {
			continue
		}

		volumes, err := client.ListVolumes(context.Background(), app.Name)
		if err != nil {
			fmt.Printf("Error listing volumes for app %s: %v\n", app.Name, err)
			continue
		}

		for _, volume := range volumes {
			fmt.Printf("Sweeping volume %s in app %s\n", volume.ID, app.Name)
			if err := client.DeleteVolume(context.Background(), app.Name, volume.ID); err != nil {
				fmt.Printf("Error sweeping volume %s: %v\n", volume.ID, err)
			}
		}
	}

	return nil
}

type sweepIP struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func sweepIPAddresses(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for IP address sweep: %w", err)
	}

	for _, app := range apps {
		if !strings.HasPrefix(app.Name, "tf-test-") {
			continue
		}

		out, err := runFlyctl("ips", "list", "-a", app.Name, "--json")
		if err != nil {
			fmt.Printf("Error listing IP addresses for app %s: %v\n", app.Name, err)
			continue
		}

		var ips []sweepIP
		if err := json.Unmarshal(out, &ips); err != nil {
			fmt.Printf("Error parsing IP addresses for app %s: %v\n", app.Name, err)
			continue
		}

		for _, ip := range ips {
			fmt.Printf("Sweeping IP address %s in app %s\n", ip.Address, app.Name)
			if _, err := runFlyctl("ips", "release", ip.Address, "-a", app.Name); err != nil {
				fmt.Printf("Error sweeping IP address %s: %v\n", ip.Address, err)
			}
		}
	}

	return nil
}

func sweepCertificates(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for certificate sweep: %w", err)
	}

	for _, app := range apps {
		if !strings.HasPrefix(app.Name, "tf-test-") {
			continue
		}

		certs, err := client.ListCertificates(context.Background(), app.Name)
		if err != nil {
			fmt.Printf("Error listing certificates for app %s: %v\n", app.Name, err)
			continue
		}

		for _, cert := range certs {
			fmt.Printf("Sweeping certificate %s in app %s\n", cert.Hostname, app.Name)
			if err := client.DeleteCertificate(context.Background(), app.Name, cert.Hostname); err != nil {
				fmt.Printf("Error sweeping certificate %s: %v\n", cert.Hostname, err)
			}
		}
	}

	return nil
}

func sweepNetworkPolicies(_ string) error {
	client, err := sweepClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApps(context.Background(), sweepOrg())
	if err != nil {
		return fmt.Errorf("listing apps for network policy sweep: %w", err)
	}

	for _, app := range apps {
		if !strings.HasPrefix(app.Name, "tf-test-") {
			continue
		}

		policies, err := client.ListNetworkPolicies(context.Background(), app.Name)
		if err != nil {
			fmt.Printf("Error listing network policies for app %s: %v\n", app.Name, err)
			continue
		}

		for _, policy := range policies {
			fmt.Printf("Sweeping network policy %s in app %s\n", policy.ID, app.Name)
			if err := client.DeleteNetworkPolicy(context.Background(), app.Name, policy.ID); err != nil {
				fmt.Printf("Error sweeping network policy %s: %v\n", policy.ID, err)
			}
		}
	}

	return nil
}

type sweepWGPeer struct {
	Name string `json:"name"`
}

func sweepWireGuardPeers(_ string) error {
	out, err := runFlyctl("wireguard", "list", sweepOrg(), "--json")
	if err != nil {
		fmt.Println("Skipping WireGuard peer sweep (list not available)")
		return nil
	}

	var peers []sweepWGPeer
	if err := json.Unmarshal(out, &peers); err != nil {
		return nil
	}

	for _, peer := range peers {
		if strings.HasPrefix(peer.Name, "tf-test-") {
			fmt.Printf("Sweeping WireGuard peer: %s\n", peer.Name)
			if _, err := runFlyctl("wireguard", "remove", sweepOrg(), peer.Name); err != nil {
				fmt.Printf("Error sweeping WireGuard peer %s: %v\n", peer.Name, err)
			}
		}
	}

	return nil
}

type sweepWGToken struct {
	Name string `json:"name"`
}

func sweepWireGuardTokens(_ string) error {
	out, err := runFlyctl("wireguard", "token", "list", sweepOrg(), "--json")
	if err != nil {
		fmt.Println("Skipping WireGuard token sweep (list not available)")
		return nil
	}

	var tokens []sweepWGToken
	if err := json.Unmarshal(out, &tokens); err != nil {
		return nil
	}

	for _, token := range tokens {
		if strings.HasPrefix(token.Name, "tf-test-") {
			fmt.Printf("Sweeping WireGuard token: %s\n", token.Name)
			if _, err := runFlyctl("wireguard", "token", "delete", sweepOrg(), "name:"+token.Name); err != nil {
				fmt.Printf("Error sweeping WireGuard token %s: %v\n", token.Name, err)
			}
		}
	}

	return nil
}

type sweepNamedResource struct {
	Name string `json:"name"`
}

type sweepOrgResource struct {
	Slug string `json:"slug"`
}

func sweepMPGClusters(_ string) error {
	out, err := runFlyctl("mpg", "list", "--json")
	if err != nil {
		fmt.Println("Skipping MPG cluster sweep (list not available)")
		return nil
	}

	var clusters []sweepNamedResource
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil
	}

	for _, cluster := range clusters {
		if strings.HasPrefix(cluster.Name, "tf-test-") {
			fmt.Printf("Sweeping MPG cluster: %s\n", cluster.Name)
			if _, err := runFlyctl("mpg", "destroy", cluster.Name, "--yes"); err != nil {
				fmt.Printf("Error sweeping MPG cluster %s: %v\n", cluster.Name, err)
			}
		}
	}

	return nil
}

func sweepRedis(_ string) error {
	out, err := runFlyctl("redis", "list", "--json")
	if err != nil {
		fmt.Println("Skipping Redis sweep (list not available)")
		return nil
	}

	var instances []sweepNamedResource
	if err := json.Unmarshal(out, &instances); err != nil {
		return fmt.Errorf("parsing Redis instances: %w", err)
	}

	for _, instance := range instances {
		if strings.HasPrefix(instance.Name, "tf-test-") {
			fmt.Printf("Sweeping Redis instance: %s\n", instance.Name)
			if _, err := runFlyctl("redis", "destroy", instance.Name, "--yes"); err != nil {
				fmt.Printf("Error sweeping Redis instance %s: %v\n", instance.Name, err)
			}
		}
	}

	return nil
}

func sweepTigrisBuckets(_ string) error {
	out, err := runFlyctl("storage", "list", "--json")
	if err != nil {
		fmt.Println("Skipping Tigris bucket sweep (list not available)")
		return nil
	}

	var buckets []sweepNamedResource
	if err := json.Unmarshal(out, &buckets); err != nil {
		return nil
	}

	for _, bucket := range buckets {
		if strings.HasPrefix(bucket.Name, "tf-test-") {
			fmt.Printf("Sweeping Tigris bucket: %s\n", bucket.Name)
			if _, err := runFlyctl("storage", "destroy", bucket.Name, "--yes"); err != nil {
				fmt.Printf("Error sweeping Tigris bucket %s: %v\n", bucket.Name, err)
			}
		}
	}

	return nil
}

func sweepPostgresClusters(_ string) error {
	out, err := runFlyctl("postgres", "list", "--json")
	if err != nil {
		// postgres list may fail or return non-JSON — skip silently.
		fmt.Println("Skipping Postgres cluster sweep (list not available)")
		return nil
	}

	var clusters []sweepNamedResource
	if err := json.Unmarshal(out, &clusters); err != nil {
		// Non-JSON output (e.g., "No postgres clusters found") — nothing to sweep.
		return nil
	}

	for _, cluster := range clusters {
		if strings.HasPrefix(cluster.Name, "tf-test-") {
			fmt.Printf("Sweeping Postgres cluster: %s\n", cluster.Name)
			if _, err := runFlyctl("postgres", "destroy", cluster.Name, "--yes"); err != nil {
				fmt.Printf("Error sweeping Postgres cluster %s: %v\n", cluster.Name, err)
			}
		}
	}

	return nil
}

func sweepOrgs(_ string) error {
	out, err := runFlyctl("orgs", "list", "--json")
	if err != nil {
		fmt.Println("Skipping org sweep (list not available)")
		return nil
	}

	var orgs []sweepOrgResource
	if err := json.Unmarshal(out, &orgs); err != nil {
		return nil
	}

	for _, org := range orgs {
		if strings.HasPrefix(org.Slug, "tf-test-") {
			fmt.Printf("Sweeping org: %s\n", org.Slug)
			if _, err := runFlyctl("orgs", "delete", org.Slug, "--yes"); err != nil {
				fmt.Printf("Error sweeping org %s: %v\n", org.Slug, err)
			}
		}
	}

	return nil
}

func sweepLiteFSClusters(_ string) error {
	out, err := runFlyctl("litefs-cloud", "clusters", "list", "--json")
	if err != nil {
		fmt.Println("Skipping LiteFS cluster sweep (list not available)")
		return nil
	}

	var clusters []sweepNamedResource
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil
	}

	for _, cluster := range clusters {
		if strings.HasPrefix(cluster.Name, "tf-test-") {
			fmt.Printf("Sweeping LiteFS cluster: %s\n", cluster.Name)
			if _, err := runFlyctl("litefs-cloud", "clusters", "destroy", cluster.Name, "--yes"); err != nil {
				fmt.Printf("Error sweeping LiteFS cluster %s: %v\n", cluster.Name, err)
			}
		}
	}

	return nil
}

func sweepExtMySQL(_ string) error {
	return makeExtSweeper("mysql")("")
}

func sweepExtSentry(_ string) error {
	return makeExtSweeper("sentry")("")
}

// makeExtSweeper returns a sweeper function for a given extension type.
// Tolerates extensions that don't support the list command.
func makeExtSweeper(extType string) func(string) error {
	return func(_ string) error {
		out, err := runFlyctl("ext", extType, "list", "--json")
		if err != nil {
			// Some extensions don't support list — skip silently.
			fmt.Printf("Skipping %s extension sweep (list not supported)\n", extType)
			return nil
		}

		var exts []sweepNamedResource
		if err := json.Unmarshal(out, &exts); err != nil {
			return fmt.Errorf("parsing %s extensions: %w", extType, err)
		}

		for _, ext := range exts {
			if strings.HasPrefix(ext.Name, "tf-test-") {
				fmt.Printf("Sweeping %s extension: %s\n", extType, ext.Name)
				if _, err := runFlyctl("ext", extType, "destroy", ext.Name, "--yes"); err != nil {
					fmt.Printf("Error sweeping %s extension %s: %v\n", extType, ext.Name, err)
				}
			}
		}

		return nil
	}
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
