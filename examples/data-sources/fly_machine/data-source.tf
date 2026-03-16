data "fly_machine" "example" {
  app = "my-existing-app"
  id  = "abc123"
}

output "machine_state" {
  value = data.fly_machine.example.state
}
