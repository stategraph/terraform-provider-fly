data "fly_organization" "org" {
  slug = "my-org"
}

output "org_name" {
  value = data.fly_organization.org.name
}
