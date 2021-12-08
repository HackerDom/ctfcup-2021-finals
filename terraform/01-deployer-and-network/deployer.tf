variable "service_account_id" {
  default = ""
}

locals {
  deployer_ip = cidrhost(module.c.jury_subnet, 5)
}

data "yandex_iam_service_account" "deployer-final" {
  name = "deployer-final"
}


resource "yandex_compute_instance" "deployer" {
  name = "deployer"
  hostname = "deployer"
  platform_id = "standard-v2"

  service_account_id = data.yandex_iam_service_account.deployer-final.id

  resources {
    cores  = 4
    memory = 8
    core_fraction = 20
  }

  boot_disk {
    auto_delete = false

    initialize_params {
      image_id = module.c.ubuntu-with-docker.id
      size = 100
      type = "network-hdd"
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.subnet.id
    ip_address = local.deployer_ip
    nat       = true
  }

  metadata = {
    ssh-keys = "ubuntu:${module.c.jury_ssh_key}"
    user-data = file("${path.module}/templates/setup_deployer.sh")
  }
}


locals {
  deployer_fip = yandex_compute_instance.deployer.network_interface.0.nat_ip_address
}

output "deployer_fip" {
  value = local.deployer_fip
}

resource "local_file" "deployer_fip" {
  content = local.deployer_fip
  filename = "${module.c.teams_path}/deployer_fip"
}

