package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	cloudConfigCluster,
	clusterType,
	cloudConfigAgent string
	serverCount int
)

func main() {

	var clp CloudProvider

	clusterType = strings.ToLower(clusterType)

	switch {
	case "amazon":
		clp = &amazon{}
	case "do":
		clp = &docean{}
	case "brightbox":
	}

	clp.initProvider()

	clp.login()

	cloudConfigCluster = createCloudConfigCluster()

	privateKey, publicKey := createSshKey()
	cloudConfigAgent = createCloudConfigAgent(publicKey)

	clusterServers := clp.createCoreOSServerCluster(serverCount, cloudConfigCluster)
	pmxAgent := clp.createCoreOSServerAgent(cloudConfigAgent)

	fleetIP := clusterServers[0].PrivateIP
	agentIP := pmxAgent.PublicIP

	setKey("agent-pri-ssh-key", base64.StdEncoding.EncodeToString([]byte(privateKey)))
	setKey("agent-fleet-api", fleetIP)
	setKey("agent-public-ip", agentIP)

	fmt.Scanln()
	time.Sleep(2000 * time.Hour)
}

func init() {
	serverCount, _ = strconv.Atoi(os.Getenv("NODE_COUNT"))
	clusterType = os.Getenv("CLUSTER_HOST")

	if serverCount == 0 {
		panic("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
	}
}
