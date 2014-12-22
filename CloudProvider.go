package main

type CloudProvider interface {
	initProvider() bool
	login() bool
	createCoreOSServerCluster(count int, cloudConfig string) []Server
	createCoreOSServerAgent(cloudConfig string) Server
	logout() bool
}

type Server struct {
	PublicIP  string
	PrivateIP string
	Name      string
}
