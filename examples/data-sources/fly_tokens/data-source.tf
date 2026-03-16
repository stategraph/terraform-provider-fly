data "fly_tokens" "app_tokens" {
  app = "my-example-app"
}

output "token_names" {
  value = [for t in data.fly_tokens.app_tokens.tokens : t.name]
}
