resource "fly_org_member" "example" {
  org   = fly_org.example.slug
  email = "teammate@example.com"
}
