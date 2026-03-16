# Terraform Provider for Fly.io

Manage your [Fly.io](https://fly.io) infrastructure with Terraform. Apps, machines, volumes, secrets, Postgres, Redis, Tigris, networking, and more.

```hcl
terraform {
  required_providers {
    fly = {
      source  = "stategraph/fly"
      version = "~> 0.1"
    }
  }
}

provider "fly" {}
```

## Quick start

You need a Fly.io API token and (for some resources) the `flyctl` CLI installed.

```bash
export FLY_API_TOKEN="fo1_your_token_here"
```

Then write some Terraform:

```hcl
resource "fly_app" "web" {
  name     = "my-web-app"
  org_slug = "personal"
}

resource "fly_machine" "web" {
  app    = fly_app.web.name
  region = "iad"
  image  = "registry.fly.io/my-web-app:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  service {
    protocol      = "tcp"
    internal_port = 8080

    port {
      port     = 443
      handlers = ["tls", "http"]
    }
  }
}

resource "fly_volume" "data" {
  app     = fly_app.web.name
  name    = "data"
  region  = "iad"
  size_gb = 10
}
```

## How it works

The provider talks to Fly.io through two backends:

**REST API** for core infrastructure — apps, machines, volumes, secrets, certificates, network policies, and snapshots. These go directly to `api.machines.dev` with rate limiting and automatic retry.

**flyctl CLI** for everything else — IPs, WireGuard, orgs, managed Postgres, Redis, Tigris, extensions, tokens. The provider shells out to `flyctl` with `--json` and parses the output. You need flyctl installed for these resources.

## Resources

### Core (REST API)

| Resource | What it does |
|----------|-------------|
| `fly_app` | Create and manage apps |
| `fly_machine` | Fly Machines (microVMs) with services, mounts, health checks |
| `fly_volume` | Persistent storage volumes (supports resize, snapshots, encryption) |
| `fly_secret` | App secrets (write-only, tracked via SHA-256 digest) |
| `fly_certificate` | TLS certificates for custom domains |
| `fly_network_policy` | Firewall rules |
| `fly_volume_snapshot` | Point-in-time volume snapshots |

### Networking (flyctl)

| Resource | What it does |
|----------|-------------|
| `fly_ip_address` | Allocate v4, v6, shared_v4, or private_v6 addresses |
| `fly_egress_ip` | Static egress IP for outbound traffic |
| `fly_wireguard_peer` | WireGuard VPN peers |
| `fly_wireguard_token` | Delegated WireGuard tokens |

### Databases (flyctl)

| Resource | What it does |
|----------|-------------|
| `fly_mpg_cluster` | Managed Postgres cluster |
| `fly_mpg_database` | Database within an MPG cluster |
| `fly_mpg_user` | User within an MPG cluster |
| `fly_mpg_attachment` | Attach MPG cluster to an app (creates `DATABASE_URL` secret) |
| `fly_postgres_cluster` | Self-managed Postgres (legacy) |
| `fly_postgres_attachment` | Attach self-managed Postgres to an app |
| `fly_redis` | Upstash Redis |

### Other (flyctl)

| Resource | What it does |
|----------|-------------|
| `fly_tigris_bucket` | Tigris object storage |
| `fly_org` | Organizations |
| `fly_org_member` | Org membership invites |
| `fly_token` | API tokens (deploy, org, etc.) |
| `fly_litefs_cluster` | LiteFS Cloud clusters |
| `fly_ext_mysql` | PlanetScale MySQL extension |
| `fly_ext_kubernetes` | Kubernetes extension |
| `fly_ext_sentry` | Sentry extension |
| `fly_ext_arcjet` | Arcjet extension |
| `fly_ext_wafris` | Wafris WAF extension |
| `fly_ext_vector` | Vector observability extension |

## Data sources

18 data sources for reading existing infrastructure:

```hcl
data "fly_app" "existing"    { name = "my-app" }
data "fly_regions" "all"     {}
data "fly_machines" "web"    { app = "my-app" }
data "fly_volumes" "storage" { app = "my-app" }
```

Available: `fly_app`, `fly_apps`, `fly_machine`, `fly_machines`, `fly_volume`, `fly_volumes`, `fly_certificate`, `fly_certificates`, `fly_volume_snapshots`, `fly_network_policies`, `fly_oidc_token`, `fly_ip_addresses`, `fly_organization`, `fly_regions`, `fly_mpg_clusters`, `fly_redis_instances`, `fly_tigris_buckets`, `fly_tokens`.

## Importing existing resources

Every resource supports `terraform import`. The import ID format varies by resource:

```bash
terraform import fly_app.web my-app-name
terraform import fly_machine.web my-app/machine-id
terraform import fly_volume.data my-app/vol_abc123
terraform import fly_secret.db_url my-app/DATABASE_URL
terraform import fly_certificate.tls my-app/app.example.com
terraform import fly_mpg_cluster.db my-pg-cluster
terraform import fly_ip_address.v6 my-app/ip-id
```

## Provider configuration

| Attribute | Env var | Description |
|-----------|---------|-------------|
| `api_token` | `FLY_API_TOKEN` | API token (required) |
| `api_url` | `FLY_API_URL` | API base URL (default: `https://api.machines.dev/v1`) |
| `org_slug` | `FLY_ORG` | Default organization |
| `flyctl_path` | `FLYCTL_PATH` | Path to flyctl binary (default: search PATH) |

Generate a token: `fly tokens create org`

## Development

### Requirements

- Go >= 1.25
- Terraform >= 1.11
- flyctl (for Layer 2 resources and acceptance tests)

### Build and install

```bash
make build     # compile the provider binary
make install   # install to local Terraform plugin dir
make lint      # run golangci-lint
make generate  # regenerate docs with tfplugindocs
```

### Testing

**Mock tests** run without credentials using `httptest` servers and a mock flyctl:

```bash
make test
```

**Acceptance tests** hit the real Fly.io API:

```bash
export FLY_API_TOKEN="fo1_..."
make testacc

# or run a single test
TF_ACC=1 go test ./internal/resources/ -run TestAccApp_basic -v -timeout 30m
```

**Sweepers** clean up leaked `tf-test-*` resources from interrupted test runs:

```bash
make sweep
```

### Releasing

Tag and push — GitHub Actions handles the rest (GoReleaser builds binaries, signs checksums with GPG, creates a GitHub release):

```bash
git tag v0.1.0
git push origin v0.1.0
```

The Terraform Registry picks up new releases automatically once the provider is registered.

## License

MIT
