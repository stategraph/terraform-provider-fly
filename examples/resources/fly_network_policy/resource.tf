resource "fly_network_policy" "allow_web" {
  app    = fly_app.example.name
  name   = "allow-web-traffic"
  action = "allow"

  source {
    app = "frontend-app"
  }

  dest {
    app   = fly_app.example.name
    ports = "8080"
    proto = "tcp"
  }
}
