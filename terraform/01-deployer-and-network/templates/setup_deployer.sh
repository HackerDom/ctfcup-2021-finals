#!/bin/bash

echo "setting up terraform and stuff"
curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add -
apt-add-repository -y "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get install -y terraform packer

echo "running vpn"
docker run --name dockovpn \
    --detach \
    --restart always \
    --cap-add=NET_ADMIN \
    -p 1194:1194/udp \
    -e HOST_ADDR=$(curl -s https://api.ipify.org) \
    --volume /srv/openvpn_conf:/doc/Dockovpn \
    alekslitvinenk/openvpn \
&& sleep 1 && docker exec dockovpn wget -O /doc/Dockovpn/client.ovpn localhost:8080

echo "you should setup yc yourself =("
