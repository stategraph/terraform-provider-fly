# Terraform Provider for Fly.io

A Terraform provider for managing [Fly.io](https://fly.io) infrastructure. Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) (Protocol v6).

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

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.11.0
- [Go](https://golang.org/doc/install) >= 1.25.0 (to build the provider)
- A [Fly.io account](https://fly.io) with an API token
- [flyctl](https://fly.io/docs/flyctl/install/) (required for managed databases, extensions, IPs, WireGuard, and other Layer 2 resources)

## Authentication

The provider authenticates via a Fly.io API token. Set it in one of two ways:

```bash
# Environment variable (recommended)
export FLY_API_TOKEN="fo1_your_token_here"
```

```hcl
# Or in provider configuration (not recommended for version control)
provider "fly" {
  api_token = var.fly_api_token
}
```

Generate a token with the [Fly.io CLI](https://fly.io/docs/flyctl/):

```bash
fly tokens create org
```

### Provider Configuration

| Attribute     | Environment Variable | Description                                      |
|---------------|---------------------|--------------------------------------------------|
| `api_token`   | `FLY_API_TOKEN`     | Fly.io API token (required)                      |
| `api_url`     | `FLY_API_URL`       | Machines API base URL (default: `https://api.machines.dev/v1`) |
| `org_slug`    | `FLY_ORG`           | Default organization slug                        |
| `flyctl_path` | `FLYCTL_PATH`       | Path to flyctl binary (default: search PATH)     |

---

## Architecture

The provider uses two layers to cover the full Fly.io platform:

- **Layer 1 (REST API):** Direct HTTP calls to `api.machines.dev/v1` for apps, machines, volumes, secrets, certificates, network policies, and volume snapshots.
- **Layer 2 (flyctl CLI):** Shells out to `flyctl` for IPs, WireGuard, organizations, regions, managed Postgres, Redis, Tigris, extensions, tokens, and orgs.

---

## Resources

The provider manages 29 resources:

### Core Infrastructure (Layer 1 — REST API)

#### `fly_app`

Creates and manages a Fly.io application.

```hcl
resource "fly_app" "example" {
  name     = "my-example-app"
  org_slug = "personal"
}
```

| Attribute  | Type   | Required | Description                          |
|------------|--------|----------|--------------------------------------|
| `name`     | string | yes      | Application name (forces replacement on change) |
| `org_slug` | string | yes      | Organization slug (forces replacement on change) |
| `id`       | string | computed | Application ID                       |
| `network`  | string | computed | Network name                         |
| `app_url`  | string | computed | Application URL                      |
| `status`   | string | computed | Application status                   |

**Import:** `terraform import fly_app.example my-app-name`

#### `fly_machine`

Creates and manages Fly.io Machines (microVMs).

```hcl
resource "fly_machine" "example" {
  app    = fly_app.example.name
  region = "iad"
  image  = "registry.fly.io/my-app:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  service {
    protocol      = "tcp"
    internal_port = 8080

    port {
      port     = 80
      handlers = ["http"]
    }

    port {
      port     = 443
      handlers = ["tls", "http"]
    }
  }

  env = {
    PORT = "8080"
  }
}
```

Key attributes: `app`, `region`, `image`, `name`, `env`, `cmd`, `entrypoint`, `auto_destroy`, `desired_status`, `cordoned`, `schedule`, `metadata`, `skip_launch`.

Blocks: `guest` (cpu_kind, cpus, memory_mb, gpu_kind, gpus), `service` (protocol, internal_port, autostart, autostop, min_machines_running, force_https, ports), `mount` (volume, path, extend_threshold_percent, add_size_gb, size_gb_limit), `metrics` (port, path), `restart` (policy, max_retries), `check` (name, type, port, interval, timeout, path, method).

Computed: `id`, `instance_id`, `state`, `private_ip`, `created_at`, `updated_at`.

**Import:** `terraform import fly_machine.example my-app/machine-id`

#### `fly_volume`

Creates and manages persistent storage volumes.

```hcl
resource "fly_volume" "data" {
  app     = fly_app.example.name
  name    = "data"
  region  = "iad"
  size_gb = 10
}
```

| Attribute                | Type   | Required | Description                           |
|--------------------------|--------|----------|---------------------------------------|
| `app`                    | string | yes      | Application name                      |
| `name`                   | string | yes      | Volume name (forces replacement)      |
| `region`                 | string | yes      | Region code (forces replacement)      |
| `size_gb`                | int    | yes      | Size in GB (can increase, not decrease) |
| `encrypted`              | bool   | optional | Enable encryption (default: true)     |
| `snapshot_id`            | string | optional | Source snapshot ID (forces replacement) |
| `snapshot_retention`     | int    | optional | Snapshot retention count              |
| `auto_backup_enabled`    | bool   | optional | Enable automatic backups              |
| `require_unique_zone`    | bool   | optional | Require unique hardware zone          |
| `state`                  | string | computed | Volume state                          |
| `attached_machine_id`    | string | computed | Attached machine ID                   |

**Import:** `terraform import fly_volume.data my-app/vol_abc123`

#### `fly_secret`

Manages application secrets. Secret values are write-only — Terraform tracks changes via a SHA-256 digest.

```hcl
resource "fly_secret" "database_url" {
  app   = fly_app.example.name
  key   = "DATABASE_URL"
  value = "postgres://user:pass@host:5432/db"
}
```

**Import:** `terraform import fly_secret.database_url my-app/DATABASE_URL`

#### `fly_certificate`

Manages TLS certificates for custom domains.

```hcl
resource "fly_certificate" "example" {
  app      = fly_app.example.name
  hostname = "app.example.com"
}
```

Computed attributes include `check_status`, `dns_validation_hostname`, `dns_validation_target`, `source`, `certificate_authority`.

**Import:** `terraform import fly_certificate.example my-app/app.example.com`

#### `fly_network_policy`

Manages network policies (firewall rules) for an application.

```hcl
resource "fly_network_policy" "allow_https" {
  app  = fly_app.example.name
  name = "allow-https"

  selector {
    all = true
  }

  rule {
    action    = "allow"
    direction = "egress"

    port {
      protocol = "tcp"
      port     = 443
    }
  }
}
```

**Import:** `terraform import fly_network_policy.allow_https my-app/policy-id`

#### `fly_volume_snapshot`

Creates an immutable snapshot of a volume. Snapshots cannot be updated or deleted (they expire automatically).

```hcl
resource "fly_volume_snapshot" "backup" {
  app       = fly_app.example.name
  volume_id = fly_volume.data.id
}
```

**Import:** `terraform import fly_volume_snapshot.backup my-app/vol-id/snap-id`

### Networking (Layer 2 — flyctl)

#### `fly_ip_address`

Allocates IP addresses for an application.

```hcl
resource "fly_ip_address" "v6" {
  app  = fly_app.example.name
  type = "v6"
}

resource "fly_ip_address" "shared_v4" {
  app  = fly_app.example.name
  type = "shared_v4"
}
```

Types: `v4`, `v6`, `shared_v4`, `private_v6`. All attributes force replacement on change.

**Import:** `terraform import fly_ip_address.v6 my-app/ip-id`

#### `fly_egress_ip`

Allocates a static egress IP address for an application. May require a paid plan.

```hcl
resource "fly_egress_ip" "example" {
  app = fly_app.example.name
}
```

**Import:** `terraform import fly_egress_ip.example my-app/ip-id`

#### `fly_wireguard_peer`

Creates a WireGuard VPN peer for an organization. All attributes force replacement.

```hcl
resource "fly_wireguard_peer" "example" {
  org_slug   = "personal"
  region     = "iad"
  name       = "my-peer"
  public_key = "BASE64_ENCODED_PUBLIC_KEY"
}
```

Computed: `peer_ip`, `endpoint_ip`, `gateway_ip`.

**Import:** `terraform import fly_wireguard_peer.example personal/my-peer`

#### `fly_wireguard_token`

Creates a delegated WireGuard token. The `token` value is only available at creation time.

```hcl
resource "fly_wireguard_token" "example" {
  org_slug = "personal"
  name     = "my-token"
}
```

**Import:** `terraform import fly_wireguard_token.example personal/my-token`

### Managed Postgres (Layer 2 — flyctl)

#### `fly_mpg_cluster`

Manages a Fly.io Managed Postgres cluster.

```hcl
resource "fly_mpg_cluster" "db" {
  name   = "my-pg-cluster"
  org    = "personal"
  region = "iad"
}
```

| Attribute          | Type   | Required | Description                              |
|--------------------|--------|----------|------------------------------------------|
| `name`             | string | yes      | Cluster name (forces replacement)        |
| `org`              | string | yes      | Organization slug (forces replacement)   |
| `region`           | string | yes      | Primary region (forces replacement)      |
| `plan`             | string | optional | Plan (free, starter, standard)           |
| `volume_size`      | int    | optional | Volume size in GB                        |
| `pg_major_version` | int    | optional | PostgreSQL major version                 |
| `enable_postgis`   | bool   | optional | Enable PostGIS extension                 |
| `status`           | string | computed | Cluster status                           |

**Import:** `terraform import fly_mpg_cluster.db my-pg-cluster`

#### `fly_mpg_database`

Creates a database within an MPG cluster.

```hcl
resource "fly_mpg_database" "mydb" {
  cluster_id = fly_mpg_cluster.db.id
  name       = "mydb"
}
```

**Import:** `terraform import fly_mpg_database.mydb cluster-id/mydb`

#### `fly_mpg_user`

Manages a user within an MPG cluster.

```hcl
resource "fly_mpg_user" "app" {
  cluster_id = fly_mpg_cluster.db.id
  username   = "app_user"
}
```

**Import:** `terraform import fly_mpg_user.app cluster-id/app_user`

#### `fly_mpg_attachment`

Attaches an MPG cluster to an application, creating a `DATABASE_URL` secret.

```hcl
resource "fly_mpg_attachment" "app" {
  cluster_id = fly_mpg_cluster.db.id
  app        = fly_app.example.name
}
```

**Import:** `terraform import fly_mpg_attachment.app cluster-id/my-app`

### Unmanaged Postgres (Layer 2 — flyctl)

#### `fly_postgres_cluster`

Creates a self-managed Postgres cluster (Fly.io Postgres app).

```hcl
resource "fly_postgres_cluster" "legacy" {
  name   = "my-postgres"
  org    = "personal"
  region = "iad"
}
```

**Import:** `terraform import fly_postgres_cluster.legacy my-postgres`

#### `fly_postgres_attachment`

Attaches an unmanaged Postgres app to another application.

```hcl
resource "fly_postgres_attachment" "app" {
  postgres_app = fly_postgres_cluster.legacy.name
  app          = fly_app.example.name
}
```

**Import:** `terraform import fly_postgres_attachment.app my-postgres/my-app`

### Redis (Layer 2 — flyctl)

#### `fly_redis`

Manages a Fly.io Upstash Redis instance.

```hcl
resource "fly_redis" "cache" {
  name   = "my-redis"
  org    = "personal"
  region = "iad"
}
```

**Import:** `terraform import fly_redis.cache my-redis`

### Tigris Object Storage (Layer 2 — flyctl)

#### `fly_tigris_bucket`

Manages a Tigris object storage bucket.

```hcl
resource "fly_tigris_bucket" "assets" {
  name = "my-assets-bucket"
  org  = "personal"
}
```

**Import:** `terraform import fly_tigris_bucket.assets my-assets-bucket`

### Organizations (Layer 2 — flyctl)

#### `fly_org`

Creates and manages a Fly.io organization.

```hcl
resource "fly_org" "team" {
  name = "my-team"
}
```

**Import:** `terraform import fly_org.team my-team-slug`

#### `fly_org_member`

Invites a member to an organization.

```hcl
resource "fly_org_member" "dev" {
  org   = fly_org.team.slug
  email = "dev@example.com"
}
```

**Import:** `terraform import fly_org_member.dev my-org/dev@example.com`

### Tokens (Layer 2 — flyctl)

#### `fly_token`

Creates Fly.io API tokens. The `token` value is only available at creation time.

```hcl
resource "fly_token" "deploy" {
  type = "deploy"
  app  = fly_app.example.name
  name = "ci-deploy"
}
```

### LiteFS Cloud (Layer 2 — flyctl)

#### `fly_litefs_cluster`

Manages a LiteFS Cloud cluster.

```hcl
resource "fly_litefs_cluster" "db" {
  name   = "my-litefs"
  org    = "personal"
  region = "iad"
}
```

**Import:** `terraform import fly_litefs_cluster.db my-litefs`

### Extensions (Layer 2 — flyctl)

Extensions are managed via a generic factory. All share the same CRUD pattern.

| Resource             | Description               | Key Attributes       |
|----------------------|---------------------------|----------------------|
| `fly_ext_mysql`      | PlanetScale MySQL         | name, org, region    |
| `fly_ext_kubernetes` | Kubernetes cluster        | name, org, region    |
| `fly_ext_sentry`     | Sentry error tracking     | name, app            |
| `fly_ext_arcjet`     | Arcjet security           | name, app            |
| `fly_ext_wafris`     | Wafris WAF                | name, app            |
| `fly_ext_vector`     | Vector observability      | name, app            |

```hcl
resource "fly_ext_sentry" "monitoring" {
  name = "my-sentry"
  app  = fly_app.example.name
}
```

**Import (all extensions):** `terraform import fly_ext_<type>.<name> resource-name`

---

## Data Sources

The provider includes 18 data sources:

| Data Source              | Lookup Key         | Layer   | Description                        |
|--------------------------|--------------------|---------|------------------------------------|
| `fly_app`                | `name`             | REST    | Single app by name                 |
| `fly_apps`               | (none)             | REST    | List all apps in account/org       |
| `fly_machine`            | `app`, `id`        | REST    | Single machine by app + ID         |
| `fly_machines`           | `app`              | REST    | List machines in an app            |
| `fly_volume`             | `app`, `id`        | REST    | Single volume by app + ID          |
| `fly_volumes`            | `app`              | REST    | List volumes in an app             |
| `fly_certificate`        | `app`, `hostname`  | REST    | Single certificate by hostname     |
| `fly_certificates`       | `app`              | REST    | List certificates for an app       |
| `fly_volume_snapshots`   | `app`, `volume_id` | REST    | List snapshots for a volume        |
| `fly_network_policies`   | `app`              | REST    | List network policies for an app   |
| `fly_oidc_token`         | `aud`              | REST    | Generate OIDC token (Fly infra only) |
| `fly_ip_addresses`       | `app`              | flyctl  | List IP addresses for an app       |
| `fly_organization`       | `slug`             | flyctl  | Organization info by slug          |
| `fly_regions`            | (none)             | flyctl  | List all available regions         |
| `fly_mpg_clusters`       | (none)             | flyctl  | List all MPG clusters              |
| `fly_redis_instances`    | (none)             | flyctl  | List all Redis instances           |
| `fly_tigris_buckets`     | (none)             | flyctl  | List all Tigris buckets            |
| `fly_tokens`             | `app` or `org`     | flyctl  | List tokens for an app or org      |

### Example

```hcl
data "fly_organization" "personal" {
  slug = "personal"
}

data "fly_regions" "all" {}

data "fly_app" "existing" {
  name = "my-existing-app"
}

data "fly_mpg_clusters" "all" {}
```

---

## Building the Provider

```bash
# Build
make build

# Install to local Terraform plugin directory
make install

# Lint
make lint
```

After `make install`, configure Terraform to use the local build:

```hcl
terraform {
  required_providers {
    fly = {
      source  = "stategraph/fly"
      version = "0.1.0"
    }
  }
}
```

---

## Testing

The test suite has three tiers: unit/mock tests, acceptance tests, and sweepers.

### Test Architecture

```
tests/
├── Unit/Mock tests      (*_mock_test.go)    — Fast, no API calls, use httptest or mock flyctl
├── Acceptance tests      (*_test.go)         — Real API calls, need FLY_API_TOKEN + flyctl
│   ├── Resource tests    (internal/resources/)
│   ├── Data source tests (internal/datasources/)
│   └── Composed tests    (composed_acc_test.go)
└── Sweepers              (sweep_test.go)     — Clean up leaked test resources
```

**Mock tests** use `httptest.NewServer` for REST resources and a mock flyctl shell script for CLI-based resources. They run fast, require no credentials, and cover full CRUD lifecycles, error handling, and drift detection.

**Acceptance tests** make real API calls against Fly.io. They create actual resources, verify they exist, test import, and confirm cleanup via `CheckDestroy`. All test resources use the `tf-test-` prefix for easy identification and cleanup.

**Sweepers** are cleanup routines that find and delete any leftover `tf-test-*` resources, with dependency ordering (machines before volumes before apps).

### Running Tests Locally

#### Mock/Unit Tests (no credentials needed)

```bash
make test
```

Runs all tests. Mock tests execute; acceptance tests are skipped automatically without `TF_ACC`.

#### Acceptance Tests (requires Fly.io API token + flyctl)

```bash
# Set your API token
export FLY_API_TOKEN="fo1_your_token_here"

# Run all acceptance tests
make testacc
```

This runs with `-parallel 2` to avoid rate-limit issues, and a 30-minute timeout.

To run a specific test:

```bash
# Single test
TF_ACC=1 go test ./internal/resources/ -v -run TestAccAppResource_basic -timeout 30m

# All tests for a specific resource
TF_ACC=1 go test ./internal/resources/ -v -run TestAccMPGCluster -timeout 30m

# All data source tests
TF_ACC=1 go test ./internal/datasources/ -v -run TestAcc -timeout 30m
```

#### Sweepers (clean up leaked resources)

If acceptance tests are interrupted or fail, resources may be left behind. Run the sweepers to clean up:

```bash
export FLY_API_TOKEN="fo1_your_token_here"
make sweep
```

This deletes all `tf-test-*` resources in dependency order. Sweepers cover apps, machines, volumes, IPs, certificates, network policies, WireGuard peers/tokens, MPG clusters, Postgres clusters, Redis instances, Tigris buckets, LiteFS clusters, orgs, and extensions.

### Test Coverage Summary

| Category                      | Count  | Description                                |
|-------------------------------|--------|--------------------------------------------|
| Resource mock tests           | 25     | httptest + mock flyctl, no credentials     |
| Data source mock tests        | 18     | httptest + mock flyctl, no credentials     |
| Resource acceptance tests     | ~40    | Real API: create, import, disappears       |
| Data source acceptance tests  | 18     | Real API: reads via data sources           |
| Composed workflow tests       | 3      | Multi-resource integration scenarios       |
| Sweepers                      | 20     | Cleanup for all resource types             |

---

## CI/CD

### GitHub Actions Workflows

#### Test Workflow (`.github/workflows/test.yml`)

Triggers on pushes to `main` and all pull requests. Three jobs:

1. **lint** — Runs `golangci-lint` on every push and PR.
2. **unit** — Runs `go test ./... -v -count=1` (mock tests) on every push and PR.
3. **acceptance** — Runs acceptance tests against real Fly.io API. **Only runs on `main`** (after merge), not on PRs, to avoid consuming API resources on every PR.

#### Required Secrets

| Secret          | Description                                | Used By     |
|-----------------|--------------------------------------------|-------------|
| `FLY_API_TOKEN` | Fly.io API token for acceptance tests      | Test workflow |
| `GPG_PRIVATE_KEY` | GPG key for signing releases             | Release workflow |
| `GPG_PASSPHRASE`  | GPG key passphrase                       | Release workflow |

To set up acceptance tests in your fork:

1. Create a Fly.io API token: `fly tokens create org`
2. Go to **Settings > Secrets and variables > Actions** in your GitHub repo
3. Add `FLY_API_TOKEN` as a repository secret

#### Release Workflow (`.github/workflows/release.yml`)

Triggers on version tags (`v*`). Uses GoReleaser to build binaries for linux/darwin/windows (amd64/arm64), sign checksums with GPG, and create a GitHub release. The Terraform Registry picks up releases automatically.

To release a new version:

```bash
git tag v0.1.0
git push origin v0.1.0
```

### Running CI Locally

You can replicate the CI pipeline locally:

```bash
# 1. Lint (requires golangci-lint)
make lint

# 2. Unit tests
make test

# 3. Acceptance tests (requires FLY_API_TOKEN + flyctl)
make testacc

# 4. Sweep cleanup
make sweep

# 5. Verify build
make build
```

---

## Project Structure

```
.
├── main.go                          # Provider entry point
├── GNUmakefile                      # Build, test, sweep targets
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider configuration and registration
│   │   ├── provider_data.go         # ProviderData struct (APIClient + Flyctl)
│   │   ├── testing.go               # Test utilities (factories, precheck, API client)
│   │   └── sweep_test.go            # Resource sweepers for cleanup (20 sweepers)
│   ├── resources/                   # 29 resource implementations
│   │   ├── app_resource.go          # Layer 1 (REST)
│   │   ├── machine_resource.go
│   │   ├── volume_resource.go
│   │   ├── secret_resource.go
│   │   ├── certificate_resource.go
│   │   ├── network_policy_resource.go
│   │   ├── volume_snapshot_resource.go
│   │   ├── ip_address_resource.go   # Layer 2 (flyctl)
│   │   ├── egress_ip_resource.go
│   │   ├── wireguard_peer_resource.go
│   │   ├── wireguard_token_resource.go
│   │   ├── mpg_cluster_resource.go
│   │   ├── mpg_database_resource.go
│   │   ├── mpg_user_resource.go
│   │   ├── mpg_attachment_resource.go
│   │   ├── postgres_cluster_resource.go
│   │   ├── postgres_attachment_resource.go
│   │   ├── redis_resource.go
│   │   ├── tigris_bucket_resource.go
│   │   ├── token_resource.go
│   │   ├── org_resource.go
│   │   ├── org_member_resource.go
│   │   ├── litefs_cluster_resource.go
│   │   ├── extension_resource.go    # Generic extension factory
│   │   ├── extension_constructors.go # 6 extension constructors
│   │   ├── *_mock_test.go           # Mock/unit tests per resource
│   │   ├── *_resource_test.go       # Acceptance tests per resource
│   │   ├── acc_test_helpers_test.go  # Shared CheckDestroy/CheckExists
│   │   └── composed_acc_test.go     # Multi-resource workflow tests
│   ├── datasources/                 # 18 data source implementations
│   │   ├── *_data_source.go
│   │   ├── datasource_mock_test.go  # Mock tests for all data sources
│   │   └── *_acc_test.go            # Acceptance tests per data source
│   ├── models/                      # Terraform state models + ProviderData
│   └── validators/                  # Custom validators
├── pkg/
│   ├── apiclient/                   # Layer 1: REST API client
│   │   ├── client.go                # HTTP client with rate limiting + retry
│   │   ├── apps.go                  # App CRUD
│   │   ├── machines.go              # Machine lifecycle
│   │   ├── volumes.go               # Volume CRUD + snapshots
│   │   ├── secrets.go               # Secret management
│   │   ├── certificates.go          # Certificate management
│   │   ├── network_policies.go      # Network policy CRUD
│   │   ├── tokens.go                # OIDC token requests
│   │   └── errors.go                # API error handling (IsNotFound, IsConflict)
│   ├── apimodels/                   # API request/response structs
│   └── flyctl/                      # Layer 2: flyctl CLI executor
│       ├── executor.go              # Command execution with mutex + timeouts
│       ├── errors.go                # FlyctlError, IsNotFound, IsAlreadyExists
│       ├── mock.go                  # MockRunner for testing
│       └── executor_test.go         # Unit tests
└── examples/                        # Example Terraform configurations
    ├── provider/
    ├── resources/
    └── data-sources/
```

### API Client (Layer 1)

The REST API client (`pkg/apiclient`) communicates with the Fly.io Machines API (`https://api.machines.dev/v1`):

- Rate limiting: 10 requests/second with burst of 10
- Automatic retry with exponential backoff for 429 and 5xx errors (3 attempts)
- Bearer token authentication

### flyctl Executor (Layer 2)

The flyctl executor (`pkg/flyctl`) wraps the `flyctl` CLI binary:

- Mutex-serialized command execution
- Auto-appends `--json` for structured output (via `RunJSON`)
- 2-minute default timeout, 10-minute timeout for creates
- `FLY_API_TOKEN` injected as environment variable
- Mockable `CommandRunner` interface for testing

---

## Publishing to Registries

### Terraform Registry

The provider is configured for Terraform Registry publishing. To publish:

1. **Set up GitHub secrets** on the `stategraph/terraform-provider-fly` repository:
   - `GPG_PRIVATE_KEY` — GPG private key for signing releases
   - `GPG_PASSPHRASE` — passphrase for the GPG key

2. **Create the first release:**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
   The release workflow (`.github/workflows/release.yml`) will build binaries, sign checksums, and create a GitHub release.

3. **Register at [registry.terraform.io](https://registry.terraform.io):**
   - Sign in with the `stategraph` GitHub account
   - Click **Publish** → **Provider**
   - Select the `terraform-provider-fly` repository
   - The registry automatically detects the `v0.1.0` release

### OpenTofu Registry

The provider is fully compatible with [OpenTofu](https://opentofu.org) — it uses Terraform Plugin Protocol v6, which OpenTofu supports natively. Once published to the Terraform Registry, OpenTofu can use it directly.

To also list on the [OpenTofu Registry](https://github.com/opentofu/registry):

1. Fork https://github.com/opentofu/registry
2. Add a provider entry at `providers/s/stategraph/fly.json`:
   ```json
   {
     "repository": "https://github.com/stategraph/terraform-provider-fly",
     "key_id": "<GPG_KEY_ID>"
   }
   ```
3. Add the GPG public key at `keys/<GPG_KEY_ID>.asc`
4. Submit a PR to the `opentofu/registry` repo

### Usage (Terraform and OpenTofu)

```hcl
terraform {
  required_providers {
    fly = {
      source  = "stategraph/fly"
      version = "~> 0.1"
    }
  }
}
```

This configuration works with both `terraform` and `tofu` CLIs.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
