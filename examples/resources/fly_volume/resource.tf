resource "fly_app" "example" {
  name     = "my-volume-app"
  org_slug = "personal"
}

resource "fly_volume" "data" {
  app    = fly_app.example.name
  name   = "data"
  region = "iad"
  size_gb = 10
}

resource "fly_machine" "example" {
  app    = fly_app.example.name
  region = "iad"
  image  = "registry.fly.io/my-app:latest"

  mount {
    volume = fly_volume.data.id
    path   = "/data"
  }

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }
}
