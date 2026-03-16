data "fly_ip_addresses" "example" {
  app = "my-existing-app"
}

output "addresses" {
  value = data.fly_ip_addresses.example.ip_addresses
}
