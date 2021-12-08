#!/bin/bash

for (( i = 101; i < 105; i += 1 )) ; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ctfcup@team$i-main:/srv/openvpn_conf/client.ovpn $i/
done
