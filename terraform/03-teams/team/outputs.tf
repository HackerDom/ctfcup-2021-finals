output "main_fip" {
  value = yandex_compute_instance.main.network_interface.0.nat_ip_address
}

output "main_id" {
  value = yandex_compute_instance.main.id
}

output "service_ips" {
  value = yandex_compute_instance.service[*].network_interface.0.ip_address
}
