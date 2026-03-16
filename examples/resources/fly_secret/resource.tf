resource "fly_app" "example" {
  name     = "my-secret-app"
  org_slug = "personal"
}

resource "fly_secret" "database_url" {
  app   = fly_app.example.name
  key   = "DATABASE_URL"
  value = "postgres://user:pass@host:5432/db"
}
