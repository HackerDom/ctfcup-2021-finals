#!/bin/bash

SERVICE_DIR="/srv/petuh/"

mkdir -p $SERVICE_DIR
cat > $SERVICE_DIR/docker-compose.yaml <<EOF
version: '2.2'

services:
  redis:
    image: redis:5-alpine
    restart: always
    pids_limit: 100

  service:
    image: ${team_registry}/qdb:latest
    stop_grace_period: 5s
    ulimits:
      nofile:
        soft: 12000
        hard: 12000
    pids_limit: 20
    command: gunicorn --chdir /server api:app -b :16962 --worker-class aiohttp.worker.GunicornWebWorker --access-logfile -
    ports:
      - "16962:16962"
    links:
      - redis

EOF

systemctl enable ctfcup@petuh
systemctl start ctfcup@petuh

