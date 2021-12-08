### teams stuff

variable "team_count" {
  default = 4
}

variable "services" {
  default = [
    {
      "name": "trash-factory",
      "template": "setup_empty.sh",
      "id": 1,
    },
    {
      "name": "5g_shop",
      "template": "setup_empty.sh",
      "id": 2,
    },
  ]
}

module "teams" {
  source = "./team"

  count = var.team_count

  subnet_id = module.c.subnet_id
  team_subnet = cidrsubnet(module.c.base_subnet, 8, 101 + count.index)
  jury_subnet = module.c.jury_subnet
  base_image_id = module.c.ubuntu-with-docker.id
  instance_prefix = "team${101 + count.index}"

  ssh_keys = [module.c.jury_ssh_key, file("${module.c.teams_path}/${101 + count.index}/ssh_key.pub")]
  serial_ssh_key = file("${module.c.teams_path}/${101 + count.index}/ssh_key.pub")

  services = var.services
}

resource "local_file" "team_ssh_host" {
  count = var.team_count
  content = "ctfcup@${module.teams[count.index].main_fip}"
  filename = "${module.c.teams_path}/${101 + count.index}/main_ssh_host"
}

### The Ansible inventory file
resource "local_file" "AnsibleInventory" {
  content = templatefile("${path.module}/templates/inventory.tmpl", {
    services = {for i in range(length(var.services)): var.services[i].name => module.teams[*].service_ips[i]}
  })
  filename = "../inventory/teams"
}

output "team_ids" {
  value = module.teams[*].main_id
}

output "team_fip" {
  value = module.teams[*].main_fip
}

