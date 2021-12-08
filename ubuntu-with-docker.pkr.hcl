
source "yandex" "ubuntu-with-docker" {
  image_description   = "Updated system with docker, docker-compose and minimal set of utils"
  image_family        = "ubuntu-with-docker"
  image_name          = "ubuntu-with-docker-{{timestamp}}"

  source_image_family = "ubuntu-2004-lts"
  disk_type           = "network-ssd"
  zone                = "ru-central1-a"
  folder_id           = "b1gqbgd717du7hpdjk5a"
  subnet_id           = "e9bmofj21dc41puiih8l"
  use_ipv4_nat        = true

  ssh_username        = "ubuntu"
}

build {
  sources = ["source.yandex.ubuntu-with-docker"]

  provisioner "shell" {
    # list of other workaround of dpkg lock: https://joelvasallo.com/?p=544
    inline = [
        "sudo apt-get update -y",
        "while PID=$(pidof -s apt-get); do tail --pid=$PID -f /dev/null; done",
        "sudo apt-get upgrade -y",
        "sudo apt-get install -y docker.io docker-compose htop atop mc tmux rsync git vim silversearcher-ag"
    ]
  }
}
