data "fly_volume_snapshots" "vol" {
  app       = "my-app"
  volume_id = "vol_abc123"
}
