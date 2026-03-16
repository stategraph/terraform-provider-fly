resource "fly_token" "deploy" {
  type   = "deploy"
  app    = fly_app.example.name
  name   = "ci-deploy-token"
  expiry = "720h"
}
