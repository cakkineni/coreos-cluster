docker build -t cakk/bbox .
docker run --rm -it -e ETCD_API=172.17.42.1:4001 -e CLUSTER_HOST=brightbox -e NODE_COUNT=1 -e CLIENT_ID=cli-b9y03 -e "CLIENT_SECRET=8qs8gamifytwbaw"  -e DATA_CENTER=ap-northeast-1 -e IMAGE=1 -e SERVER_SIZE=t1.micro cakk/bbox:latest
