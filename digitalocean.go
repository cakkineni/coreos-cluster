package main

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"github.com/digitalocean/godo"
	"os"
	"strconv"
)

var (
	location,
	keyName,
	apiToken string
	doClient *godo.Client
	sshKeyID []interface{}
)

type docean struct {
}

func (do *docean) initProvider() bool {
	apiToken = os.Getenv("DIGITALOCEAN_TOKEN")
	location = os.Getenv("REGION")
	keyName := os.Getenv("SSH_KEY_NAME")
	size = os.Getenv("VM_SIZE")

	if apiToken == "" || location == "" || keyName == "" || size == "" {
		panic("\n\nMissing Params...Check Docs...\n\n")
	}
	return true
}

func (do *docean) login() {

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: apiToken},
	}

	doClient = godo.NewClient(t.Client())

	intIds := []int{getSSHKeyID()}
	for _, v := range intIds {
		sshKeyID = append(sshKeyID, v)
	}
}

func getSSHKeyID() int {
	keys, _, err := doClient.Keys.List(&godo.ListOptions{Page: 1, PerPage: 10})
	keyID := -1

	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if key.Name == keyName {
			keyID = key.ID
			break
		}
	}

	if keyID == -1 {
		panic(fmt.Sprintf("\n\nSSH Key Name not found. Please make sure it matches exactly to your setup (case sensitive)\n\n"))
	}

	return keyID
}

func (do *docean) createCoreOSServerCluster(count int, cloudConfig string) []*Server {

	var createReq *godo.DropletCreateRequest
	createReq = &godo.DropletCreateRequest{
		Region:            location,
		Size:              size,
		Image:             "coreos-stable",
		PrivateNetworking: true,
		UserData:          cloudConfig,
		SSHKeys:           sshKeyID,
	}

	var coreOSCluster Server
	for i := 0; i < count; i++ {
		createReq.Name = "coreos-" + strconv.Itoa(i+1)
		coreOSCluster = createServer(createReq)
	}

	return nil
}

func (do *docean) createCoreOSServerAgent(cloudConfig string) *Server {
	println("Create CoreOS Agent Server")
	var createReq *godo.DropletCreateRequest
	createReq = &godo.DropletCreateRequest{
		Name:              "pmx-remote-agent",
		Region:            location,
		Size:              "512mb",
		Image:             "coreos-stable",
		PrivateNetworking: true,
		UserData:          cloudConfig,
		SSHKeys:           sshKeyID,
	}
	server := createServer(createReq)
	return &Server{Name: server.Droplet.Name, PublicIP: server.Droplet.Networks.V4[1].IPAddress, PrivateIP: server.Droplet.Networks.V4[0].IPAddress}
}

func createServer(createRequest *godo.DropletCreateRequest) *godo.DropletRoot {
	var err error
	newDroplet, _, err := doClient.Droplets.Create(createRequest)

	if err != nil {
		panic(err)
	}
	return newDroplet
}

func deleteServer(id int) {
	doClient.Droplets.Delete(id)
}
