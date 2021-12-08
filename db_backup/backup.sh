#!/bin/bash

host=cs-main
db=cs
interval_sec=60

while true; do
    echo "Running backup at $(date)"
    time ssh "$host" "sudo -u postgres -- pg_dump '$db' | gzip" > "backup_$(date +%Y-%m-%d-%H-%M-%S).gz";
    sleep 60;
done
