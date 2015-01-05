#!/bin/bash
clear
docker build -t cakk/pmxdeploy .

#docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=brightbox -e NODE_COUNT=1 -e CLIENT_ID=****** -e CLIENT_SECRET=******* -e REGION=zon-qt3kk -e VM_SIZE=nano cakk/pmxdeploy:latest
#docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=docean -e NODE_COUNT=1 -e DIGITALOCEAN_TOKEN=******* -e REGION=nyc3 -e "SSH_KEY_NAME=macbook air" -e VM_SIZE=512MB cakk/pmxdeploy:latest
#docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=amazon -e NODE_COUNT=1 -e AWS_ACCESS_KEY_ID=***** -e AWS_SECRET_ACCESS_KEY=****8 -e REGION=us-east-1b -e "SSH_KEY_NAME=mac" -e VM_SIZE=t1.micro cakk/pmxdeploy:latest
docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=centurylink -e NODE_COUNT=1  -e "API_KEY=********" -e "API_PASSWORD=********" -e REGION=UC1 -e VM_SIZE=nano -e "NETWORK_NAME=CoreOS Test DHCP Network" cakk/pmxdeploy:latest
