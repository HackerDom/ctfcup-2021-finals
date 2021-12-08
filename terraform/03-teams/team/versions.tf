terraform {
  required_providers {
    cloudinit = {
      source = "hashicorp/cloudinit"
    }
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
  required_version = ">= 0.13"
}
