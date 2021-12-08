#!/bin/bash

for (( i = 101; i < 105; i += 1 )) ; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null TEAM_README ctfcup@10.118.$i.10:README
done
