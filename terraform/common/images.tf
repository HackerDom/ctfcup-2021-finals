data "yandex_resourcemanager_folder" "ctf" {
  name = "ctfcup2021-final"
}

data "yandex_compute_image" "ubuntu-with-docker" {
  family = "ubuntu-with-docker"
  folder_id = data.yandex_resourcemanager_folder.ctf.id
}

output "ubuntu-with-docker" {
  value = data.yandex_compute_image.ubuntu-with-docker
}

