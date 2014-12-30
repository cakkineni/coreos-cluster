FROM progrium/busybox
RUN opkg-install bash
RUN mkdir -p /etc/ssl && mkdir -p /etc/ssl/certs
ADD certs /etc/ssl/certs/
ADD cluster/cloud-config-agent.yaml cloud-config-agent.yaml
ADD cluster/cloud-config-init.yaml cloud-config-init.yaml
ENV SHELL /bin/bash
ADD cluster/cluster cluster
RUN chmod +x cluster
CMD "./cluster"
#docker run -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=brightbox -e NODE_COUNT=1 -e CLIENT_ID=cli-q07lz   -e "CLIENT_SECRET=23mzyqo1pi2nrko"  -e DATA_CENTER=ap-northeast-1 -e IMAGE=1 -e SERVER_SIZE=t1.micro cakk/cluster:latest
