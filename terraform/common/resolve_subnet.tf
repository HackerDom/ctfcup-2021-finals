variable "resolve_subnet" {
  default = true
}

data "yandex_vpc_subnet" "subnet" {
  count = var.resolve_subnet ? 1 : 0
  name = "ctf-subnet"
}

output "subnet_id" {
  value = join("", data.yandex_vpc_subnet.subnet[*].id)
}
