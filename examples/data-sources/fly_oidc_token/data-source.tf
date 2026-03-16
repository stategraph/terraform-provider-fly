data "fly_oidc_token" "github" {
  app = "my-app"
  aud = "https://github.com/my-org"
}
