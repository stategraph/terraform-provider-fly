terraform {
  required_providers {
    fly = {
      source  = "stategraph/fly"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.11.0"
}

provider "fly" {
  # api_token  = var.fly_api_token  # Or set FLY_API_TOKEN env var
  # org_slug   = "personal"         # Or set FLY_ORG env var
  # flyctl_path = "/usr/local/bin/flyctl"  # Path to flyctl binary (auto-detected if omitted)
}
