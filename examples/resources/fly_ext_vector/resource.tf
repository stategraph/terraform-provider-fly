resource "fly_ext_vector" "logs" {
  name = "my-vector"
  app  = fly_app.example.name
}
