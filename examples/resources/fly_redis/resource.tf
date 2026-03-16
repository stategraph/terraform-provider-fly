resource "fly_redis" "example" {
  name            = "my-redis"
  org             = "personal"
  region          = "iad"
  plan            = "launch-1"
  enable_eviction = true
  replica_regions = ["ord"]
}
