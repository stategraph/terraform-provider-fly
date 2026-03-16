resource "fly_ext_arcjet" "security" {
  name = "my-arcjet"
  app  = fly_app.example.name
}
