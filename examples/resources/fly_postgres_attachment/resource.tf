resource "fly_postgres_attachment" "example" {
  postgres_app  = fly_postgres_cluster.example.name
  app           = fly_app.example.name
  database_name = "my_app_db"
  variable_name = "DATABASE_URL"
}
