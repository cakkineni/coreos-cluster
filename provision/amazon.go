package provision

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"os"
	"time"
)

//brew install bazaar
//go get launchpad.net/goamz/ec2

type Amazon struct {
	location  string
	keyName   string
	amiName   string
	size      string
	amzClient *ec2.EC2
}

func NewAmazon() *Amazon {
	return new(Amazon)
}

func (amz Amazon) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	println("\nProvisioning Cluster in Amazon EC2")
	pmxCluster := PMXCluster{}
	println("\nInitializing...")
	amz.initProvider()
	println("\nLogging into EC2...")
	amz.login()
	println("\nProvisioning CoreOS cluster")
	pmxCluster.Cluster = amz.provisionCoreOSCluster(params.ServerCount, params.CloudConfigCluster)
	println("\Deploying Panamax Remote Agent server")
	agent := amz.provisionPMXAgent(params.CloudConfigAgent)
	pmxCluster.Agent = agent
	println("\nLogging out...")
	amz.logout()
	return pmxCluster
}

func (amz *Amazon) initProvider() bool {
	apiToken := os.Getenv("AWS_ACCESS_KEY_ID")
	apiPassword := os.Getenv("AWS_SECRET_ACCESS_KEY")
	amz.location = os.Getenv("REGION")
	amz.keyName = os.Getenv("SSH_KEY_NAME")
	amz.size = os.Getenv("VM_SIZE")
	amz.amiName = os.Getenv("AMI_NAME")

	if amz.amiName == "" {
		amz.amiName = "ami-f469f29c"
	}

	if apiToken == "" || apiPassword == "" || amz.location == "" || amz.size == "" || amz.amiName == "" {
		panic("\n\nMissing Values. Ensure you have all environment variables needed to create cluster...\n\n")
	}
	return true
}

func (amz *Amazon) login() bool {
	auth, err := aws.EnvAuth()

	if err != nil {
		panic(err)
	}

	amz.amzClient = ec2.New(auth, aws.USEast)
	return true
}

func (amz *Amazon) logout() bool {
	return true
}

func (amz *Amazon) provisionCoreOSCluster(count int, cloudConfig string) []Server {
	createReq := &ec2.RunInstances{
		ImageId:      amz.amiName,
		InstanceType: amz.size,
		UserData:     []byte(cloudConfig),
		MinCount:     count,
		MaxCount:     count,
		KeyName:      amz.keyName,
		AvailZone:    amz.location,
	}

	var coreOSServers []Server
	for i := 0; i < count; i++ {
		println("Provisioning Server ", i+1)
		coreOSServers = append(coreOSServers, amz.createServer(createReq))
	}
	return coreOSServers
}

func (amz *Amazon) provisionPMXAgent(cloudConfig string) Server {
	createReq := ec2.RunInstances{
		ImageId:      amz.amiName,
		InstanceType: amz.size,
		UserData:     []byte(cloudConfig),
		MinCount:     1,
		MaxCount:     1,
		KeyName:      amz.keyName,
		AvailZone:    amz.location,
	}

	return amz.createServer(&createReq)
}

func (amz *Amazon) createServer(createRequest *ec2.RunInstances) Server {
	var resp *ec2.RunInstancesResp
	resp, err := amz.amzClient.RunInstances(createRequest)

	if err != nil {
		panic(err.Error())
	}

	server := resp.Instances[0]
	println("Waiting for server creation to be complete...")
	for {
		print(".")
		time.Sleep(10 * time.Second)
		resp, err := amz.amzClient.Instances([]string{server.InstanceId}, &ec2.Filter{})

		if err != nil {
			panic(err)
		}

		if server.State.Code == 16 {
			break
		}

		server = resp.Reservations[0].Instances[0]
	}
	return Server{PublicIP: server.IPAddress, Name: server.DNSName, PrivateIP: server.PrivateIPAddress}
}
