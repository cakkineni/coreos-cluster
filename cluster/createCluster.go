package main

import (
	"encoding/base64"
	"fmt"
	provision "github.com/cakkineni/coreos-cluster/provision"
	"os"
	"strconv"
	"time"
)

func main() {
	println("Deploying PMX Cluster")
	clusterHost, nodeCount, targetName := readParams()
	clp := provision.New(clusterHost)

	println("Generating Cloud Config")
	cloudConfigCluster := createCloudConfigCluster()
	privateKey, publicKey := createSSHKey()
	cloudConfigAgent := createCloudConfigAgent(publicKey)

	clusterParams := provision.ClusterParams{}
	clusterParams.ServerCount = nodeCount
	clusterParams.CloudConfigAgent = cloudConfigAgent
	clusterParams.CloudConfigCluster = cloudConfigCluster

	println("Provisioning PMX Cluster")
	cluster := clp.ProvisionPMXCluster(clusterParams)
	fleetIP := cluster.Cluster[0].PrivateIP
	agentIP := cluster.Agent.PublicIP

	setKey("AGENT_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(privateKey)))
	setKey("AGENT_FLEET_API", fleetIP)
	setKey("AGENT_PUB_IP", agentIP)
	setKey("REMOTE_TARGET_NAME", targetName)

	println("Provisioning Complete!!!")
	fmt.Scanln()
	time.Sleep(2000 * time.Hour)
}

func readParams() (string, int, string) {
	nodeCount, _ := strconv.Atoi(os.Getenv("NODE_COUNT"))
	clusterHost := os.Getenv("CLUSTER_HOST")
	targetName := os.Getenv("REMOTE_TARGET_NAME")

	if nodeCount == 0 || clusterHost == "" || targetName == "" {
		panic("\n\nMissing Params...Please Check Docs...\n\n")
	}
	return clusterHost, nodeCount, targetName
}
