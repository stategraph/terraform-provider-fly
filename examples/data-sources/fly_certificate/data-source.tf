data "fly_certificate" "example" {
  app      = "my-existing-app"
  hostname = "example.com"
}

output "cert_status" {
  value = data.fly_certificate.example.check_status
}

output "dns_target" {
  value = data.fly_certificate.example.dns_validation_target
}
