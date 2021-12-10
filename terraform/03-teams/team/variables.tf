variable "jury_subnet" {
  type = string
}
variable "team_subnet" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "base_image_id" {
  type = string
}

variable "instance_prefix" {
  type = string
}

variable "main_resources" {
  default = {
    cores = 4
    memory = 8
    core_fraction = 100
    preemtible = false
  }
}

variable "ssh_keys" {
  default = []
}

variable "serial_ssh_key" {
  default = ""
}

variable "serial_ssh_password_hash" {
  # echo 'petuh' | mkpasswd --method=SHA-512 --rounds=4096 --stdin
  default = "$6$rounds=4096$4csUlFsu$VO5THDmn3vhs9GMQ5YCmO.zp.IzeZ1qlgtoLsOY6k.V1qsEW41YlnpP7Zq1BhM4l9iVtgTQVPaFhUrQoMQrFm1"
}

variable "services" {
  default = []
  type = list(map(string))
}

locals {
  main_ip = cidrhost(var.team_subnet, 10)
}

