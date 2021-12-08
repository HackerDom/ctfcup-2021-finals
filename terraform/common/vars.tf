### Subnets
locals {
  base_subnet = "10.118.0.0/16"
}

output "base_subnet" {
  value = local.base_subnet
}

output "jury_subnet" {
  value = cidrsubnet(local.base_subnet, 8, 0)
}

### Paths and keys
locals {
  teams_path = "${path.module}/../../teams"
}

output "teams_path" {
  value = local.teams_path
}

output "jury_ssh_key" {
  value = file("${local.teams_path}/for_devs.ssh_key.pub")
}

### Notes
# 10.118.0.0/24 - jury
# 10.118.0.5 - infra stuff
# 10.118.0.10-11 - checksystem
# 10.118.0.20-21 - dns + nginx
# 10.118.101-110.0/24 - teams
# 10.118.1xx.10 - main
# 10.118.1xx.11 - s1
# 10.118.1xx.12 - s2
# 10.118.1xx.13 - s3
