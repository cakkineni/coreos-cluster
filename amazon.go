package main

import (
	"encoding/json"
	"io/ioutil"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"os"
	"time"
)

//brew install bazaar
//go get launchpad.net/goamz/ec2

var (
	location,
	keyName,
	amiName,
	size string
	amzClient *ec2.EC2
)

type amazon struct {
}

func (amz *amazon) initProvider() bool {
	apiToken := os.Getenv("AWS_ACCESS_KEY_ID")
	apiPassword := os.Getenv("AWS_SECRET_ACCESS_KEY")
	location = os.Getenv("REGION")
	keyName = os.Getenv("SSH_KEY_NAME")
	size = os.Getenv("VM_SIZE")

	var amis []struct {
		Region string
		AMI    string
	}

	amiFile, _ := ioutil.ReadFile("aws_ami.json")
	json.Unmarshal(amiFile, &amis)

	for _, ami := range amis {
		if ami.Region == location {
			amiName = ami.AMI
			break
		}
	}

	println("AMI Used:" + amiName)

	if apiToken == "" || apiPassword == "" || location == "" || size == "" || amiName == "" {
		panic("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
	}
	return true
}

func (amz *amazon) login() {
	println("\nLogging in....")
	auth, err := aws.EnvAuth()

	if err != nil {
		panic(err)
	}

	amzClient = ec2.New(auth, aws.USEast)
}

func (amz *amazon) createCoreOSServerCluster(count int, cloudConfig string) []*Server {
	println("Create CoreOS Server")
	createReq := &ec2.RunInstances{
		ImageId:      amiName,
		InstanceType: size,
		UserData:     []byte(cloudConfig),
		MinCount:     count,
		MaxCount:     count,
		KeyName:      keyName,
	}

	var coreOSCluster Server
	for i := 0; i < count; i++ {
		coreOSCluster = createServer(createReq)
	}
	return nil
}

func (amz *amazon) createCoreOSServerAgent(cloudConfig string) *Server {
	println("Create CoreOS Agent Server")
	createReq := ec2.RunInstances{
		ImageId:      amiName,
		InstanceType: size,
		UserData:     []byte(cloudConfig),
		MinCount:     1,
		MaxCount:     1,
		KeyName:      keyName,
	}

	return createServer(&createReq)
}

func createServer(createRequest *ec2.RunInstances) *Server {
	var resp *ec2.RunInstancesResp
	resp, err := amzClient.RunInstances(createRequest)

	if err != nil {
		panic(err.Error())
	}

	server := resp.Instances[0]
	for {
		println("Server State:" + server.State.Name)

		time.Sleep(10 * time.Second)
		resp, err := amzClient.Instances([]string{server.InstanceId}, &ec2.Filter{})

		if err != nil {
			panic(err)
		}

		if server.State.Code == 16 {
			break
		}

		server = resp.Reservations[0].Instances[0]
	}
	return &Server{PublicIP: server.IPAddress, Name: server.DNSName, PrivateIP: server.PrivateIPAddress}
}
