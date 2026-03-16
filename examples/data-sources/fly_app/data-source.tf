data "fly_app" "example" {
  name = "my-existing-app"
}

output "app_status" {
  value = data.fly_app.example.status
}
