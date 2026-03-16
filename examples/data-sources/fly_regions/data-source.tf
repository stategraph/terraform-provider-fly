data "fly_regions" "all" {}

output "region_codes" {
  value = [for r in data.fly_regions.all.regions : r.code]
}
