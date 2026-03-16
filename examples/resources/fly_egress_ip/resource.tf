resource "fly_egress_ip" "static" {
  app = fly_app.example.name
}
