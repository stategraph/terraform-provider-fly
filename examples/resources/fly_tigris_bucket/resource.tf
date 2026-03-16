resource "fly_tigris_bucket" "example" {
  name   = "my-storage-bucket"
  org    = "personal"
  public = false
}
