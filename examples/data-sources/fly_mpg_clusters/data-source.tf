data "fly_mpg_clusters" "all" {}

output "cluster_names" {
  value = [for c in data.fly_mpg_clusters.all.clusters : c.name]
}
