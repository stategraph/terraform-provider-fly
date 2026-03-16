resource "fly_wireguard_token" "ci" {
  org_slug = "my-org"
  name     = "ci-token"
}
