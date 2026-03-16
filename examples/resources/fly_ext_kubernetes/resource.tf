resource "fly_ext_kubernetes" "cluster" {
  name   = "my-k8s"
  org    = "personal"
  region = "iad"
}
