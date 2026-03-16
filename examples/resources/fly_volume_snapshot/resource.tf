resource "fly_volume_snapshot" "backup" {
  app       = fly_app.example.name
  volume_id = fly_volume.data.id
}
