resource "fly_ext_sentry" "example" {
  name = "my-sentry"
  app  = fly_app.example.name
}
