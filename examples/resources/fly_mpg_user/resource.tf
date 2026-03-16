resource "fly_mpg_user" "example" {
  cluster_id = fly_mpg_cluster.example.id
  username   = "app_user"
  role       = "readwrite"
}
