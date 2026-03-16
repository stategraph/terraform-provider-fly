resource "fly_mpg_database" "example" {
  cluster_id = fly_mpg_cluster.example.id
  name       = "my_database"
}
