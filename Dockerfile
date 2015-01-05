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