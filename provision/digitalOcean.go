package provision

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"github.com/digitalocean/godo"
	"os"
	"strconv"
	"time"
)

var ()

type DOcean struct {
	location string
	keyName  string
	apiToken string
	size     string
	doClient *godo.Client
	sshKeyID []interface{}
}

func NewDOcean() *DOcean {
	return new(DOcean)
}

func (doc DOcean) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	println("\nProvisioning PMX Cluster in Digital Ocean")
	cluster := PMXCluster{}
	println("\nInitializing")
	doc.initProvider()
	println("\nLogging in to Digital Ocean")
	doc.login()
	println("\nProvisioning CoreOS cluster")
	cluster.Cluster = doc.provisionCoreOSCluster(params.ServerCount, params.CloudConfigCluster)
	println("\nProvisioning Panamax Remote Agent")
	cluster.Agent = doc.provisionPMXAgent(params.CloudConfigAgent)
	println("\nLogging out")
	doc.logout()
	return cluster
}

func (doc *DOcean) initProvider() bool {
	doc.apiToken = os.Getenv("DIGITALOCEAN_TOKEN")
	doc.location = os.Getenv("REGION")
	doc.keyName = os.Getenv("SSH_KEY_NAME")
	doc.size = os.Getenv("VM_SIZE")

	if doc.apiToken == "" || doc.location == "" || doc.keyName == "" || doc.size == "" {
		panic("\n\nMissing Params...Check Docs...\n\n")
	}
	return true
}

func (doc *DOcean) login() bool {

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: doc.apiToken},
	}

	doc.doClient = godo.NewClient(t.Client())

	intIds := []int{doc.getSSHKeyID()}
	for _, v := range intIds {
		doc.sshKeyID = append(doc.sshKeyID, v)
	}
	return true
}

func (doc *DOcean) logout() bool {
	return true
}

func (doc *DOcean) getSSHKeyID() int {
	keys, _, err := doc.doClient.Keys.List(&godo.ListOptions{Page: 1, PerPage: 10})
	keyID := -1

	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if key.Name == doc.keyName {
			keyID = key.ID
			break
		}
	}

	if keyID == -1 {
		panic(fmt.Sprintf("\n\nSSH Key Name not found. Please make sure it matches exactly to your setup (case sensitive)\n\n"))
	}

	return keyID
}

func (doc *DOcean) provisionCoreOSCluster(count int, cloudConfig string) []Server {
	var createReq *godo.DropletCreateRequest
	createReq = &godo.DropletCreateRequest{
		Region:            doc.location,
		Size:              doc.size,
		Image:             "coreos-stable",
		PrivateNetworking: true,
		UserData:          cloudConfig,
		SSHKeys:           doc.sshKeyID,
	}

	var coreOSServers []Server
	for i := 0; i < count; i++ {
		println("Provisioning Server ", i+1)
		createReq.Name = "coreos-" + strconv.Itoa(i+1)
		droplet := doc.createServer(createReq)
		coreOSServers = append(coreOSServers, Server{Name: droplet.Droplet.Name, PublicIP: droplet.Droplet.Networks.V4[1].IPAddress, PrivateIP: droplet.Droplet.Networks.V4[0].IPAddress})
	}
	return coreOSServers
}

func (doc *DOcean) provisionPMXAgent(cloudConfig string) Server {
	var createReq *godo.DropletCreateRequest
	createReq = &godo.DropletCreateRequest{
		Name:              "pmx-remote-agent",
		Region:            doc.location,
		Size:              "512mb",
		Image:             "coreos-stable",
		PrivateNetworking: true,
		UserData:          cloudConfig,
		SSHKeys:           doc.sshKeyID,
	}
	server := doc.createServer(createReq)
	return Server{Name: server.Droplet.Name, PublicIP: server.Droplet.Networks.V4[1].IPAddress, PrivateIP: server.Droplet.Networks.V4[0].IPAddress}
}

func (doc *DOcean) createServer(createRequest *godo.DropletCreateRequest) *godo.DropletRoot {
	var err error
	newDroplet, _, err := doc.doClient.Droplets.Create(createRequest)

	if err != nil {
		panic(err)
	}

	println("Waiting for server creation to be complete..")
	for {
		print(".")
		newDroplet, _, _ = doc.doClient.Droplets.Get(newDroplet.Droplet.ID)
		if newDroplet.Droplet.Status == "active" {
			break
		}
		time.Sleep(60 * time.Millisecond)
	}
	return newDroplet
}

func (doc *DOcean) deleteServer(id int) {
	doc.doClient.Droplets.Delete(id)
}
