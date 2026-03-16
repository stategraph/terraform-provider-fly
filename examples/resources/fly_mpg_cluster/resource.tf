resource "fly_mpg_cluster" "example" {
  name             = "my-mpg-cluster"
  org              = "personal"
  region           = "iad"
  plan             = "launch-2"
  volume_size      = 10
  pg_major_version = 16
}
