#!/bin/bash

rm /etc/resolv.conf
cat > /etc/resolv.conf <<EOF
nameserver 8.8.4.4
nameserver 8.8.8.8

search ru-central1.internal
EOF

systemctl restart systemd-resolved.service

# configure docker and firewall

mkdir -p /etc/docker/
cat > /etc/docker/daemon.json <<EOF
{
  "insecure-registries" : ["${team_registry}"]
}
EOF
systemctl restart docker

# by the way, add ctfcup to docker group
adduser ctfcup docker

# firewall
# inspired by https://unrouted.io/2017/08/15/docker-firewall/
cat > /srv/iptables.conf <<EOF
# To reload use systemctl restart ctf-iptables
# or just iptables-restore -n /srv/iptables.conf

*filter
:INPUT ACCEPT [0:0]
:FORWARD DROP [0:0]
:OUTPUT ACCEPT [0:0]
:FILTERS - [0:0]
:DOCKER-USER - [0:0]

-F INPUT
-F DOCKER-USER
-F FILTERS

-A INPUT -i eth0 -j FILTERS
-A DOCKER-USER -i eth0 -j FILTERS


-A FILTERS -m conntrack --ctstate RELATED,ESTABLISHED -j RETURN
-A FILTERS -p tcp --dport 22 -j RETURN
-A FILTERS -p udp --dport 1194 -j RETURN
-A FILTERS -p tcp --dport 1194 -j RETURN
-A FILTERS -s ${team_subnet} -j RETURN
-A FILTERS -s ${jury_subnet} -j RETURN
-A FILTERS -j DROP

COMMIT
EOF

cat > /etc/systemd/system/ctf-iptables.service <<EOF
[Unit]
Description=Restore iptables firewall rules
Before=network-pre.target

[Service]
Type=oneshot
ExecStart=/sbin/iptables-restore -n /srv/iptables.conf

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ctf-iptables
systemctl start ctf-iptables

#ufw disable
#ufw reset --force
#ufw default deny incoming
#ufw default allow outgoing
#ufw allow 22
#ufw allow from "${team_subnet}"
#ufw allow from "${jury_subnet}"
#ufw enable
