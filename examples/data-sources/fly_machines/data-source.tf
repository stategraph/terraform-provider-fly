data "fly_machines" "all" {
  app = "my-app"
}

output "machine_ids" {
  value = [for m in data.fly_machines.all.machines : m.id]
}
