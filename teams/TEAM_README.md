# README:
 * checksystem: http://10.118.0.10
 * service proxy: `10.118.0.20`

## go to services of other teams:
 * 101 <= `<team_id>` <= 104
 * Service "trash-factory" has 2 ports:
   `10.118.0.20:1<team_id>` and `10,118.0.20:2<team_id>`
 * Service "5g_shop" has 1 port:
   `10.118.0.20:3<team_id>`

## game info
 * round time: 60 seconds
 * flag lifetime: 12 minutes
 * flag format: [A-Z0-9]{31}=

## service and deploy info
 * vms with services: 10.118.<team_subnet>.11, 10.118.<team_subnet>.12
 * firewall config: /srv/iptables.conf
 * vpn - runs in docker, client config in /srv/openvpn/, you are free to replace it with your own

## submitting flags
curl -s -H 'X-Team-Token: your_secret_token' -X PUT -d '["PNFP4DKBOV6BTYL9YFGBQ9006582ADC=", "STH5LK9R9OMGXOV4E06YZD71F746F53=", "0I7DUCYPX8UB2HP6D6UGN86BA26F2FE=", "PTK3DAGZ6XU4LPETXJTN7CE30EC0B54="]' http://10.118.0.10/flags | json_pp

 ____________
< GOOD LUCK! >
 ------------
 \
  \
   \ >()_
      (__)__ _
