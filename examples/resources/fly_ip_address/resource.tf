resource "fly_app" "example" {
  name     = "my-ip-app"
  org_slug = "personal"
}

resource "fly_ip_address" "v4" {
  app  = fly_app.example.name
  type = "shared_v4"
}

resource "fly_ip_address" "v6" {
  app  = fly_app.example.name
  type = "v6"
}
