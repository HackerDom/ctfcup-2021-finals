
data "cloudinit_config" "service" {
  count = length(var.services)
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

  # ne vzletelo
  #part {
  #  content = file("${path.module}/metadata/run_service_systemd_unit.sh")
  #}

  part {
    content = templatefile("${path.module}/metadata/${var.services[count.index].template}", {
      service_name = var.services[count.index].name,
      team_registry = "${local.main_ip}:5000"
    })
    filename = var.services[count.index].template
  }
}


resource "yandex_compute_instance" "service" {
  count = length(var.services)
  name = "${var.instance_prefix}-${var.services[count.index].name}"
  hostname = "${var.instance_prefix}-${var.services[count.index].name}"

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
      size = 80
      type = "network-ssd"
    }
  }

  network_interface {
    subnet_id = var.subnet_id
    ip_address = cidrhost(var.team_subnet, 10 + var.services[count.index].id)
    nat = false
  }

  metadata = {
    # how to connect via ssh: ssh -o ControlPath=none -o IdentitiesOnly=yes -o CheckHostIP=no -o UserKnownHostsFile=./serialssh-knownhosts -p 9600 -i ~/.ssh/id_rsa epdo9egutunl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
    # or simpler command: ssh -p 9600 epdo9egutunl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
    serial-port-enable = 1
    ssh-keys = "yc-serialssh:${var.serial_ssh_key}"
    user-data = data.cloudinit_config.service[count.index].rendered
  }
}



