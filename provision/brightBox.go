package provision

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type BrightBox struct {
	bboxApi     string
	apiKey      string
	apiPassword string
	groupName   string
	imageName   string
	serverSize  string
	letters     []rune
	httpUtil    HttpUtil
	serverCount int
	openPorts   string
}

func NewBrightBox() *BrightBox {
	cl := new(BrightBox)
	cl.bboxApi = "https://api.gb1.brightbox.com"
	cl.httpUtil = HttpUtil{APIEndPoint: cl.bboxApi}
	return cl
}

func (bbox BrightBox) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	println("\nProvisioning PMX Cluster in Brightbox")
	cluster := PMXCluster{}

	println("\nInitializing")
	bbox.initProvider()

	println("\nLogging in to Brightbox")
	loggedIn := bbox.login()

	if !loggedIn {
		panic("Login Failed!!")
	}

	println("\nProvisioning CoreOS cluster in Brightbox")
	clusterName := fmt.Sprintf("coreos-%s", bbox.randSeq(4))

	println("\nCreating server group:", clusterName)
	bbox.groupName = bbox.createGroup(clusterName)

	println("\nCreating Firewall Policy:", clusterName)
	fwPolicyId := bbox.createFWPolicy(clusterName, bbox.groupName)

	println("\nCreating Firewall Rules:", clusterName)
	bbox.createFWRules(fwPolicyId)

	println("\nProvisioning CoreOS cluster...")
	cluster.Cluster = bbox.provisionCoreOSCluster(params.ServerCount, params.CloudConfigCluster)

	println("\nProvisioning Panamax Remote Agent")
	agent := bbox.createCoreOSServer("pmx_agent", params.CloudConfigAgent, "nano")
	cluster.Agent = Server{Name: agent.Id, PrivateIP: agent.Interfaces[0].IP}
	publicIP := bbox.addPublicIP(agent.Interfaces[0].ID)
	cluster.Agent.PublicIP = publicIP

	println("Logging Out...")
	bbox.logout()

	return cluster
}

func (bbox *BrightBox) initProvider() bool {
	bbox.apiKey = os.Getenv("CLIENT_ID")
	bbox.apiPassword = os.Getenv("CLIENT_SECRET")
	bbox.imageName = os.Getenv("IMAGE")
	bbox.serverSize = os.Getenv("VM_SIZE")
	bbox.letters = []rune("abcdefghijklmnopqrstuvwxyz")
	bbox.openPorts = fmt.Sprintf("22,7001,4001,3001,%s", os.Getenv("OPEN_PORTS"))

	if bbox.imageName == "" {
		bbox.imageName = "img-kpruj"
	}

	if bbox.apiKey == "" || bbox.apiPassword == "" {
		panic("\n\nMissing Params...Check Docs....\n\n")
	}

	return true
}

func (bbox *BrightBox) login() bool {

	var postData = struct {
		ClientID  string `json:"client_id"`
		GrantType string `json:"grant_type"`
	}{bbox.apiKey, "none"}

	_, err := json.Marshal(postData)

	if err != nil {
		panic(err)
	}

	resp := bbox.httpUtil.doBasicAuth("/token", bbox.apiKey, bbox.apiPassword, postData)

	var status struct {
		AccessToken string `json:"access_token"`
	}

	json.Unmarshal([]byte(resp), &status)
	if status.AccessToken == "" {
		panic("Login Failed, Please check credentials.")
	}
	bbox.httpUtil.Headers = append(bbox.httpUtil.Headers, KeyValue{Key: "Authorization", Value: fmt.Sprintf("OAuth %s", status.AccessToken)})
	return true
}

func (bbox *BrightBox) logout() bool {
	bbox.httpUtil.HttpClient.Get(bbox.bboxApi + "/auth/logout")
	return true
}

func (bbox *BrightBox) provisionCoreOSCluster(count int, cloudConfig string) []Server {
	var coreosServers []Server
	for i := 0; i < count; i++ {
		println("Provisioning Server ", i+1)
		bboxServer := bbox.createCoreOSServer("coreos", cloudConfig, bbox.serverSize)
		fmt.Printf("%+v", bboxServer)
		coreosServers = append(coreosServers, Server{Name: bboxServer.Id, PrivateIP: bboxServer.Interfaces[0].IP})
	}
	return coreosServers
}

func (bbox *BrightBox) createGroup(groupName string) string {
	var postData = struct {
		Name string `json:"name"`
	}{groupName}
	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/server_groups", postData)
	var resp struct {
		Id string `json:"id"`
	}
	json.Unmarshal([]byte(respNewGroup), &resp)
	return resp.Id
}

func (bbox *BrightBox) createFWPolicy(policyName string, servergroupId string) string {
	var postData = struct {
		Group string `json:"server_group"`
		Name  string `json:"name"`
	}{servergroupId, policyName}
	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/firewall_policies", postData)
	var resp struct {
		Id string `json:"id"`
	}
	json.Unmarshal([]byte(respNewGroup), &resp)
	return resp.Id
}

func (bbox *BrightBox) createFWRules(firewallPolicyId string) {
	var postData = struct {
		FirewallPolicy  string `json:"firewall_policy"`
		Protocol        string `json:"protocol"`
		Source          string `json:"source"`
		DestinationPort string `json:"destination_port"`
	}{firewallPolicyId, "tcp", "any", bbox.openPorts}

	bbox.httpUtil.postJSONData("/1.0/firewall_rules", postData)
	var postData2 = struct {
		FirewallPolicy string `json:"firewall_policy"`
		Destination    string `json:"destination"`
	}{firewallPolicyId, "any"}

	bbox.httpUtil.postJSONData("/1.0/firewall_rules", postData2)
}

func (bbox *BrightBox) addPublicIP(serverID string) string {
	var postData = struct {
		Name string `json:"name"`
	}{"coreos"}
	resp := bbox.httpUtil.postJSONData("/1.0/cloud_ips", postData)
	var cloudIP struct {
		Id       string `json:"id"`
		PublicIP string `json:"public_ip"`
	}
	json.Unmarshal([]byte(resp), &cloudIP)

	var postDataMap = struct {
		Destination string `json:"destination"`
	}{serverID}

	time.Sleep(20 * time.Second)
	resp = bbox.httpUtil.postJSONData(fmt.Sprintf("/1.0/cloud_ips/%s/map", cloudIP.Id), postDataMap)
	return cloudIP.PublicIP
}

func (bbox *BrightBox) randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = bbox.letters[rand.Intn(len(bbox.letters))]
	}
	return string(b)
}

func (bbox *BrightBox) createCoreOSServer(name string, cloudConfig string, vmSize string) BBoxServer {
	b64CloudConfig := base64.StdEncoding.EncodeToString([]byte(cloudConfig))
	var postData = struct {
		Name         string   `json:"name"`
		Image        string   `json:"image"`
		ServerType   string   `json:"server_type"`
		UserData     string   `json:"user_data"`
		ServerGroups []string `json:"server_groups"`
	}{name, bbox.imageName, vmSize, b64CloudConfig, []string{bbox.groupName}}
	var respNewServer = bbox.httpUtil.postJSONData("/1.0/servers", postData)
	var resp BBoxServer
	json.Unmarshal([]byte(respNewServer), &resp)
	return resp
}

type BBoxServer struct {
	Id         string `json:"id"`
	Interfaces []struct {
		IP string `json:"ipv4_address"`
		ID string `json:"id"`
	} `json:"interfaces"`
}
