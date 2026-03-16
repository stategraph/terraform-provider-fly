resource "fly_app" "example" {
  name     = "my-machine-app"
  org_slug = "personal"
}

resource "fly_machine" "example" {
  app    = fly_app.example.name
  region = "iad"
  image  = "registry.fly.io/my-app:latest"

  guest {
    cpu_kind  = "shared"
    cpus      = 1
    memory_mb = 256
  }

  service {
    protocol      = "tcp"
    internal_port = 8080

    port {
      port     = 80
      handlers = ["http"]
    }

    port {
      port     = 443
      handlers = ["tls", "http"]
    }
  }

  env = {
    PORT = "8080"
  }
}
