resource "yandex_vpc_network" "network" {
  name = "ctf-net"
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "ctf-subnet"
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.network.id
  v4_cidr_blocks = [module.c.base_subnet]
}

