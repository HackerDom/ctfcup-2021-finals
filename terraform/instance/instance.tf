# this exists because terraform is actually pretty shit, as concept of blocks inside resources is... rather strange
# so it's quite hard to make abstraction of half-configured resources
# ... should have used polumi or written api calls myself instead...

variable "name" {
  type = string
}

variable "resources" {
  # most of args are actually numbers... but this is terraform so double-coversions are the least of my troubles
  type = map(string)
}

variable "image_id" {
  type = string
}

variable "nat"  {
  default = false
}

variable "subnet_id" {
  type = string
}

variable "ip_address" {
  type = string
}

variable "metadata" {
  type = map(string)
}

resource "yandex_compute_instance" "i" {
  name = var.name
  hostname = var.name

  platform_id = "standard-v2"

  resources {
    cores = var.resources.cores
    memory = var.resources.memory
    core_fraction = lookup(var.resources, "core_fraction", 100)
  }

  scheduling_policy {
    preemptible = lookup(var.resources, "preemptible", false)
  }

  boot_disk {
    initialize_params {
      image_id = var.image_id
      size = var.resources.disk_size
      type = lookup(var.resources, "disk_type", "network-ssd")
    }
  }

  network_interface {
    subnet_id = var.subnet_id
    ip_address = var.ip_address
    nat       = var.nat
  }

  metadata = var.metadata
}

output "id" {
  value = yandex_compute_instance.i.id
}

output "ip" {
  value = yandex_compute_instance.i.network_interface.0.ip_address
}

output "fip" {
  value = yandex_compute_instance.i.network_interface.0.nat_ip_address
}

