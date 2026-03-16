resource "fly_mpg_attachment" "example" {
  cluster_id    = fly_mpg_cluster.example.id
  app           = fly_app.example.name
  database      = fly_mpg_database.example.name
  variable_name = "DATABASE_URL"
}
