package main

import (
	"encoding/base64"
	"fmt"
	provision "github.com/cakkineni/coreos-cluster/provision"
	"os"
	"strconv"
	"time"
)

var (
	cloudConfigCluster,
	clusterType,
	cloudConfigAgent string
	serverCount int
)

func main() {

	clp := provision.New(clusterType)

	cloudConfigCluster = createCloudConfigCluster()
	privateKey, publicKey := createSSHKey()
	cloudConfigAgent = createCloudConfigAgent(publicKey)

	clusterParams := provision.ClusterParams{}
	clusterParams.ServerCount = serverCount
	clusterParams.CloudConfigAgent = cloudConfigAgent
	clusterParams.CloudConfigCluster = cloudConfigCluster

	cluster := clp.ProvisionPMXCluster(clusterParams)

	fleetIP := cluster.Cluster[0].PrivateIP
	agentIP := cluster.Agent.PublicIP

	setKey("agent-pri-ssh-key", base64.StdEncoding.EncodeToString([]byte(privateKey)))
	setKey("agent-fleet-api", fleetIP)
	setKey("agent-public-ip", agentIP)

	fmt.Scanln()
	time.Sleep(2000 * time.Hour)
}

func init() {
	serverCount, _ = strconv.Atoi(os.Getenv("NODE_COUNT"))
	clusterType = os.Getenv("CLUSTER_HOST")

	if serverCount == 0 || clusterType == "" {
		panic("\n\nMissing Params...Please Check Docs...\n\n")
	}
}