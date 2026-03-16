resource "fly_ext_wafris" "firewall" {
  name = "my-wafris"
  app  = fly_app.example.name
}
