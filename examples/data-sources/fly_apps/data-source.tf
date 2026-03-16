data "fly_apps" "all" {
  org_slug = "my-org"
}

output "app_names" {
  value = [for app in data.fly_apps.all.apps : app.name]
}
