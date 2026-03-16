data "fly_redis_instances" "all" {}

output "redis_names" {
  value = [for i in data.fly_redis_instances.all.instances : i.name]
}
