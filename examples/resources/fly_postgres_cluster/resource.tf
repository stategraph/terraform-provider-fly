resource "fly_postgres_cluster" "example" {
  name           = "my-pg-cluster"
  org            = "personal"
  region         = "iad"
  cluster_size   = 1
  volume_size    = 10
  vm_size        = "shared-cpu-1x"
  enable_backups = true
}
