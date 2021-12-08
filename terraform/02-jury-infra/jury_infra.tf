variable "dev_resources" {
  default = {
    cores = 4,
    core_fraction = 20,
    memory = 8,
    disk_size = 50,
  }
}

variable "prod_resources" {
  default = {
    cores  = 16,
    memory = 32,
    core_fraction = 100,
    disk_size = 100,
  }
}

locals {
  resources = var.dev_resources
}

module "cs-main" {
  source = "../instance"

  name = "cs-main"
  resources = local.resources
  image_id = module.c.ubuntu-with-docker.id
  subnet_id = module.c.subnet_id
  ip_address = cidrhost(module.c.jury_subnet, 10)
  metadata = {
    ssh-keys = "ubuntu:${module.c.jury_ssh_key}"
  }
}

module "cs-worker" {
  source = "../instance"

  name = "cs-worker"
  resources = local.resources
  image_id = module.c.ubuntu-with-docker.id
  subnet_id = module.c.subnet_id
  ip_address = cidrhost(module.c.jury_subnet, 11)
  metadata = {
    ssh-keys = "ubuntu:${module.c.jury_ssh_key}"
  }
}


module "proxy" {
  source = "../instance"
  count = 1

  name = "proxy-${count.index}"
  resources = local.resources
  image_id = module.c.ubuntu-with-docker.id
  subnet_id = module.c.subnet_id
  ip_address = cidrhost(module.c.jury_subnet, 20 + count.index)
  metadata = {
    ssh-keys = "ubuntu:${module.c.jury_ssh_key}"
  }
}

### The Ansible inventory file
resource "local_file" "AnsibleInventory" {
  content = templatefile("${path.module}/templates/inventory.tmpl", {
    cs-main-ip = module.cs-main.ip,
    cs-worker-ip = module.cs-worker.ip,
    proxy-ip = module.proxy[*].ip,
  })
  filename = "../inventory/jury"
}
