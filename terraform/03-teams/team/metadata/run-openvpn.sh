#!/bin/bash

# vpn`
docker run --name dockovpn \
    --detach \
    --restart always \
    --cap-add=NET_ADMIN \
    -p 1194:1194/udp \
    -e HOST_ADDR=$(curl -s https://api.ipify.org) \
    --volume /srv/openvpn_conf:/doc/Dockovpn \
    alekslitvinenk/openvpn \
&& sleep 1 && docker exec dockovpn wget -O /doc/Dockovpn/client.ovpn localhost:8080

# registry
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry \
  -v /srv/registry:/var/lib/registry \
  registry:2
