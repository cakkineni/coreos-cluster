#!/bin/sh
docker build -t cakk/pmxdeploy .
docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=docean -e NODE_COUNT=1 -e DIGITALOCEAN_TOKEN=f9150cca6bb0c5a746ee1292dde55c062c939183dfc97e70bf559be0edc17247 \
                    -e REGION=nyc3 -e "SSH_KEY_NAME=macbook air" -e VM_SIZE=512MB cakk/pmxdeploy:latest
