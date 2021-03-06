package provision

import (
	"fmt"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

//brew install bazaar
//go get launchpad.net/goamz/ec2

type Amazon struct {
	location       string
	keyName        string
	amiName        string
	size           string
	suffix         string
	openPorts      string
	securityGroups []ec2.SecurityGroup
	amzClient      *ec2.EC2
}

func NewAmazon() *Amazon {
	return new(Amazon)
}

func (amz Amazon) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	println("\nProvisioning PMX Cluster in Amazon EC2")
	pmxCluster := PMXCluster{}

	println("\nInitializing")
	amz.initProvider()

	println("\nLogging in to EC2")
	amz.login()

	println("\nProvisioning CoreOS cluster")
	amz.createFirewallRules()
	pmxCluster.Cluster = amz.provisionCoreOSCluster(params.ServerCount, params.CloudConfigCluster)

	println("\nProvisioning Panamax Remote Agent")
	agent := amz.provisionPMXAgent(params.CloudConfigAgent)
	pmxCluster.Agent = agent

	println("\nLogging out")
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
	additionalPorts := os.Getenv("OPEN_TCP_PORTS")
	defaultPorts := "22,7001,4001,30001"

	if additionalPorts != "" {
		amz.openPorts = fmt.Sprintf("%s,%s", defaultPorts, additionalPorts)
	} else {
		amz.openPorts = defaultPorts
	}

	amz.suffix = amz.randSeq(4)

	if amz.amiName == "" {
		amz.amiName = "ami-f469f29c"
	}

	if apiToken == "" || apiPassword == "" || amz.location == "" || amz.size == "" || amz.amiName == "" {
		panic("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
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

func (amz *Amazon) createFirewallRules() {
	resp, err := amz.amzClient.CreateSecurityGroup("coreos-"+amz.suffix, "CoreOS Cluster")
	if err != nil {
		panic(err)
	}
	var perms []ec2.IPPerm
	var ports []string
	ports = strings.Split(amz.openPorts, ",")
	for _, val := range ports {
		port, err := strconv.Atoi(val)
		if err == nil {
			perms = append(perms, ec2.IPPerm{Protocol: "TCP", ToPort: port, FromPort: port, SourceIPs: []string{"0.0.0.0/0"}})
		}
	}
	_, err = amz.amzClient.AuthorizeSecurityGroup(resp.SecurityGroup, perms)
	if err != nil {
		panic(err)
	}
	amz.securityGroups = append(amz.securityGroups, resp.SecurityGroup)
}

func (amz *Amazon) provisionCoreOSCluster(count int, cloudConfig string) []Server {
	createReq := &ec2.RunInstances{
		ImageId:        amz.amiName,
		InstanceType:   amz.size,
		UserData:       []byte(cloudConfig),
		MinCount:       0,
		MaxCount:       0,
		KeyName:        amz.keyName,
		AvailZone:      amz.location,
		SecurityGroups: amz.securityGroups,
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
		ImageId:        amz.amiName,
		InstanceType:   amz.size,
		UserData:       []byte(cloudConfig),
		MinCount:       0,
		MaxCount:       0,
		KeyName:        amz.keyName,
		AvailZone:      amz.location,
		SecurityGroups: amz.securityGroups,
	}

	return amz.createServer(&createReq)
}

func (amz *Amazon) createServer(createRequest *ec2.RunInstances) Server {
	resp, err := amz.amzClient.RunInstances(createRequest)
	if err != nil {
		panic(err)
	}
	server := resp.Instances[0]
	println("Waiting for server creation to be complete..")
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

func (amz *Amazon) randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
