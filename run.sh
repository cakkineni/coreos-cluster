#!/bin/bash
clear
docker rmi `docker images | grep none | awk '{print $3}'`
docker build -t cakk/pmxdeploy .

#docker run --rm -it -e "REMOTE_TARGET_NAME=BBOX" -e CLUSTER_HOST=brightbox -e NODE_COUNT=2  -e "OPEN_PORTS=8080,3306" -e CLIENT_ID=******* -e CLIENT_SECRET=******** -e VM_SIZE=nano cakk/pmxdeploy:latest
#docker run --rm -it -e "REMOTE_TARGET_NAME=DOCEAN"  -e CLUSTER_HOST=digitalocean -e NODE_COUNT=2 -e DIGITALOCEAN_TOKEN=******* -e REGION=nyc3 -e "SSH_KEY_NAME=macbook air" -e VM_SIZE=512MB cakk/pmxdeploy:latest
#docker run --rm -it -e "REMOTE_TARGET_NAME=AMAZON" -e CLUSTER_HOST=amazon -e NODE_COUNT=1 -e "OPEN_PORTS=8080,3306" -e AWS_ACCESS_KEY_ID=********* -e "AWS_SECRET_ACCESS_KEY=*********" -e REGION=us-east-1b -e "SSH_KEY_NAME=mac" -e VM_SIZE=t1.micro cakk/pmxdeploy:latest
#docker run --rm -it -e "REMOTE_TARGET_NAME=CENTURYLINK"   -e CLUSTER_HOST=centurylink -e NODE_COUNT=2  -e "API_KEY=********" -e "API_PASSWORD=********" -e REGION=UC1 -e VM_SIZE=nano -e "NETWORK_NAME=CoreOS Test DHCP Network" cakk/pmxdeploy:latest
