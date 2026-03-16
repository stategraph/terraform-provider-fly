data "fly_tigris_buckets" "all" {}

output "bucket_names" {
  value = [for b in data.fly_tigris_buckets.all.buckets : b.name]
}
