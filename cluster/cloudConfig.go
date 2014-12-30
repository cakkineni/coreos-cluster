package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func createCloudConfigCluster() string {
	println("Create Cloud Config Cluster")
	httpClient := &http.Client{}
	response, _ := httpClient.Get("https://discovery.etcd.io/new")
	defer response.Body.Close()
	contents, _ := ioutil.ReadAll(response.Body)
	cloudConfig, _ := ioutil.ReadFile("cloud-config-init.yaml")
	discoveryURL := fmt.Sprintf("discovery: %s", string(contents))
	cloudConfigNew := strings.Replace(string(cloudConfig), "discovery_url", discoveryURL, -1)
	return string(cloudConfigNew)
}

func createCloudConfigAgent(pubKey string) string {
	println("Create Cloud Config Agent")
	cloudConfig, _ := ioutil.ReadFile("cloud-config-agent.yaml")
	cloudConfigNew := strings.Replace(string(cloudConfig), "ssh-rsa", pubKey, -1)
	return string(cloudConfigNew)
}
