package provision

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type BrightBox struct {
	bboxApi     string
	location    string
	apiKey      string
	apiPassword string
	groupName   string
	imageName   string
	serverSize  string
	letters     []rune
	httpUtil    HttpUtil
	serverCount int
}

func NewBrightBox() *BrightBox {
	cl := new(BrightBox)
	cl.bboxApi = "https://api.gb1.brightbox.com"
	cl.httpUtil = NewHttpUtil()
	return cl
}

func (bbox BrightBox) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	cluster := PMXCluster{}
	bbox.initProvider()
	loggedIn := bbox.login()
	if !loggedIn {
		panic("Login Failed!!")
	}

	serverGroupId := bbox.createGroup("coreos")
	fwPolicyId := bbox.createFWPolicy("coreos", serverGroupId)
	bbox.createFWRules(fwPolicyId)

	cluster.Cluster = bbox.provisionCoreOSCluster(params.ServerCount, params.CloudConfigCluster)
	cluster.Agent = bbox.createCoreOSServer("pmx_agent", params.CloudConfigAgent)
	bbox.logout()
	return cluster
}

func (bbox *BrightBox) initProvider() bool {
	bbox.apiKey = os.Getenv("CLIENT_ID")
	bbox.apiPassword = os.Getenv("CLIENT_SECRET")
	bbox.location = os.Getenv("DATA_CENTER")
	bbox.imageName = os.Getenv("IMAGE")
	bbox.serverSize = os.Getenv("SERVER_SIZE")
	bbox.letters = []rune("abcdefghijklmnopqrstuvwxyz")

	println("Key" + bbox.apiKey + ":" + bbox.apiPassword)

	if bbox.imageName == "" {
		bbox.imageName = "coreos"
	}

	if bbox.apiKey == "" || bbox.apiPassword == "" || bbox.location == "" {
		panic("\n\nMissing Params...Check Docs....\n\n")
	}

	bbox.groupName = fmt.Sprintf("%s-%s", bbox.groupName, bbox.randSeq(4))
	bbox.bboxApi = "https://api.gb1.brightbox.com"
	println("API" + bbox.bboxApi + ":" + bbox.apiKey + ":" + bbox.apiPassword)
	bbox.httpUtil = HttpUtil{APIEndPoint: bbox.bboxApi}
	return true
}

func (bbox *BrightBox) login() bool {

	var postData = struct {
		ClientID  string `json:"client_id"`
		GrantType string `json:"grant_type"`
	}{bbox.apiKey, "none"}

	val, err := json.Marshal(postData)
	println("Logging in..." + string(val))
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
		return false
	}
	bbox.httpUtil.Headers = append(bbox.httpUtil.Headers, KeyValue{Key: "Authorization", Value: fmt.Sprintf("OAuth %s", status.AccessToken)})
	return true
}

func (bbox *BrightBox) logout() bool {
	println("\nLogging out...")
	bbox.httpUtil.HttpClient.Get(bbox.bboxApi + "/auth/logout")
	return true
}

func (bbox *BrightBox) provisionCoreOSCluster(count int, cloudConfig string) []Server {

	var coreosServers []Server
	for i := 0; i < count; i++ {
		server := bbox.createCoreOSServer("coreos", cloudConfig)
		coreosServers[i] = server
	}

	return coreosServers
}

func (bbox *BrightBox) createGroup(groupName string) string {
	fmt.Printf("\nCreating Server Group in Data Center %s with name: %s", bbox.location, bbox.groupName)

	var postData = struct {
		name string
	}{groupName}

	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/server_groups", postData)
	var resp struct {
		id string
	}
	json.Unmarshal([]byte(respNewGroup), &resp)

	println(respNewGroup)

	return resp.id
}

func (bbox *BrightBox) createFWPolicy(policyName string, servergroupId string) string {
	fmt.Printf("\nCreating Firewall Policy in Data Center %s with name: %s", bbox.location, bbox.groupName)

	var postData = struct {
		server_group string
		name         string
	}{servergroupId, policyName}

	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/server_groups", postData)
	var resp struct {
		id string
	}
	json.Unmarshal([]byte(respNewGroup), &resp)

	println(respNewGroup)

	return resp.id
}

func (bbox *BrightBox) createFWRules(firewallPolicyId string) int {
	fmt.Printf("\nCreating Firewall Rules in Data Center %s with name: %s", bbox.location, bbox.groupName)

	var postData = struct {
		firewall_policy  string
		protocol         string
		source           string
		destination_port string
		destination      string
	}{firewallPolicyId, "tcp", "any", "22,7001,4001,3001,8080,3306", "any"}

	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/firewall_rules", postData)

	println(respNewGroup)

	return 0
}

func (bbox *BrightBox) randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = bbox.letters[rand.Intn(len(bbox.letters))]
	}
	return string(b)
}

func (bbox *BrightBox) createCoreOSServer(name string, cloudConfig string) Server {
	var postData = struct {
		image         string
		name          string
		server_type   string
		zone          string
		user_data     string
		server_groups []string
	}{bbox.imageName, name, bbox.serverSize, bbox.location, cloudConfig, []string{bbox.groupName}}

	var respNewGroup = bbox.httpUtil.postJSONData("/1.0/servers", postData)

	println(respNewGroup)

	return Server{}
}
