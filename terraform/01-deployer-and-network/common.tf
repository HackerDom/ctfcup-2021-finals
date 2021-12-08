provider "yandex" {
  zone  = "ru-central1-b"
}

module "c" {
  source = "../common/"
  resolve_subnet = false
}
