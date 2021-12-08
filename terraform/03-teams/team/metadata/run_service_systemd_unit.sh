#!/bin/bash

cat > /srv/run-and-update-docker-compose.sh << EOF
#!/bin/bash
while true; do
    docker-compose pull && docker-compose up -d
    sleep 10
done
EOF

chmod a+x /srv/run-and-update-docker-compose.sh

cat > /etc/systemd/system/ctfcup@.service << EOF
[Unit]
Description=CupCTF 2020: %i
After=network-online.target docker.service
Wants=network-online.target docker.service
Requires=docker.service

[Service]
ExecStart=/srv/run-and-update-docker-compose.sh
PostExecStop=docker-compose stop
WorkingDirectory=/srv/%i
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
