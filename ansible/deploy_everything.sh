#!/bin/bash

mkdir -p .log
ansible-playbook cs-deploy.yml deploy-proxy.yml deploy-services.yml cs-init.yml cs-start.yml
