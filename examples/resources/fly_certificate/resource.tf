resource "fly_app" "example" {
  name     = "my-cert-app"
  org_slug = "personal"
}

resource "fly_certificate" "example" {
  app      = fly_app.example.name
  hostname = "example.com"
}
