resource "fly_wireguard_peer" "office" {
  org_slug   = "my-org"
  region     = "iad"
  name       = "office-vpn"
  public_key = "your-wireguard-public-key"
}
