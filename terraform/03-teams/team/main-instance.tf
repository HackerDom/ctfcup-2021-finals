data "cloudinit_config" "main" {
  gzip = false
  base64_encode = false

  part {
    content = templatefile("${path.module}/metadata/meta.txt", {
      user = "ctfcup",
      ssh_keys = var.ssh_keys,
      hashed_password = var.serial_ssh_password_hash
    })
  }

  part {
    content = templatefile("${path.module}/metadata/configure-network.sh", {
      team_subnet = var.team_subnet,
      team_registry = "${local.main_ip}:5000",
      jury_subnet = var.jury_subnet
    })
    filename = "configure-network.sh"
  }

  part {
    content = file("${path.module}/metadata/run-openvpn.sh")
    filename = "run-openvpn.sh"
  }
}

resource "yandex_vpc_address" "main-addr" {
  name = "${var.instance_prefix}-static-ip"

  external_ipv4_address {
    zone_id = "ru-central1-b"
  }
}

resource "yandex_compute_instance" "main" {
  name = "${var.instance_prefix}-main"
  hostname = "${var.instance_prefix}-main"

  platform_id = "standard-v2"

  resources {
    cores  = var.main_resources.cores
    memory = var.main_resources.memory
    core_fraction = var.main_resources.core_fraction
  }

  scheduling_policy {
    preemptible = var.main_resources.preemtible
  }

  boot_disk {
    initialize_params {
      image_id = var.base_image_id
      size = 100
      type = "network-hdd"
    }
  }

  network_interface {
    subnet_id = var.subnet_id
    ip_address = local.main_ip
    nat_ip_address = yandex_vpc_address.main-addr.external_ipv4_address[0].address
    nat = true
  }

  metadata = {
    # how to connect via ssh: ssh -o ControlPath=none -o IdentitiesOnly=yes -o CheckHostIP=no -o UserKnownHostsFile=./serialssh-knownhosts -p 9600 -i ~/.ssh/id_rsa epdo9egutunl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
    # or simpler command: ssh -p 9600 epdo9egutunl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
    serial-port-enable = 1
    ssh-keys = "yc-serialssh:${var.serial_ssh_key}"
    user-data = data.cloudinit_config.main.rendered
  }
}

