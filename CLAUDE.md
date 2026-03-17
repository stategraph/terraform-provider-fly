# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build provider binary
make install        # Build + install to local Terraform plugin dir
make test           # Unit/mock tests: go test ./... -v
make testacc        # Acceptance tests (needs FLY_API_TOKEN): TF_ACC=1 go test ./internal/... -v -parallel 2 -timeout 30m
make lint           # golangci-lint run ./...
make generate       # go generate ./... (runs tfplugindocs)
make sweep          # Clean up leaked tf-test-* resources from failed acceptance tests
```

Run a single test:
```bash
go test ./internal/resources/ -run TestAccApp_basic -v          # acceptance (needs TF_ACC=1)
go test ./internal/resources/ -run TestMockMachineResource -v   # mock (no credentials needed)
```

## Architecture

### Two-Layer API Design

The provider uses two distinct backends to talk to Fly.io:

**Layer 1 — REST API** (`pkg/apiclient/`): Direct HTTP calls to `api.machines.dev/v1` with rate limiting (10 req/s, burst 10) and retry with exponential backoff on 429/5xx. Used for: apps, machines, volumes, secrets, certificates, network policies, volume snapshots.

**Layer 2 — flyctl CLI** (`pkg/flyctl/`): Shells out to the `flyctl` binary with mutex-serialized execution (flyctl isn't concurrent-safe). Auto-appends `--json` for structured output. Used for: IPs, WireGuard, organizations, regions, managed Postgres, Redis, Tigris, extensions, tokens.

### Package Layout

- `internal/provider/` — Provider config, resource/datasource registration, test helpers, sweepers
- `internal/resources/` — 29 resource implementations (CRUD + import)
- `internal/datasources/` — 18 data source implementations (read-only)
- `internal/models/` — Terraform state model structs (`tfsdk` tags), `ProviderData` shared config
- `pkg/apiclient/` — HTTP client with auth, rate limiting, retry
- `pkg/apimodels/` — API request/response structs
- `pkg/flyctl/` — flyctl executor wrapper with `CommandRunner` interface for mocking

### Resource Implementation Pattern

Every resource implements `Resource`, `ResourceWithConfigure`, and `ResourceWithImportState`. Layer 1 resources receive `*apiclient.Client` via `Configure()`; Layer 2 resources receive `*flyctl.Executor`. Drift detection: `Read()` checks `IsNotFound()` → `resp.State.RemoveResource()`.

### Testing Tiers

1. **Mock tests** (`*_mock_test.go`): Use `httptest.NewServer` to mock the Fly API. Run without credentials. Each test covers full CRUD lifecycle via `resource.UnitTest`.
2. **Acceptance tests** (`*_test.go`): Hit real Fly.io API. Require `FLY_API_TOKEN` and `FLY_ORG` env vars. Three patterns per resource: `_basic`, `_import`, `_disappears`. All test resources use `tf-test-` prefix.
3. **Sweepers** (`provider/sweep_test.go`): Clean up leaked test resources in dependency order.

### Error Handling

- REST errors: `apiclient.IsNotFound()` / `IsConflict()` check HTTP status codes
- flyctl errors: `flyctl.IsNotFound()` / `IsAlreadyExists()` use stderr pattern matching

## Provider Configuration

| Attribute    | Env Var          | Required |
|-------------|------------------|----------|
| `api_token`  | `FLY_API_TOKEN`  | Yes      |
| `api_url`    | `FLY_API_URL`    | No       |
| `org_slug`   | `FLY_ORG`        | No       |
| `flyctl_path` | `FLYCTL_PATH`   | No       |
| `dry_run`    | `FLY_DRY_RUN`   | No       |
