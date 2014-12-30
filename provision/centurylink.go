package provision

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	//"net/http/httputil"
	//"flag"
	"math/rand"
	"os"
	"time"
)

type CenturyLink struct {
	clcApi            string
	dhcpServerAlias   string
	coreosServerAlias string
	createdDhcpName   string
	location          string
	networkName       string
	groupName         string
	apiKey            string
	apiPassword       string
	accountAlias      string
	publicIP          string
	serverPassword    string
	groupId           int
	serverCount       int
	httpUtil          HttpUtil
}

var (
	letters []rune
)

func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	cl.clcApi = "https://api.tier3.com/rest"
	cl.dhcpServerAlias = "DHCP"
	cl.coreosServerAlias = "COREOS"
	letters = []rune("abcdefghijklmnopqrstuvwxyz")
	return cl
}

func (clc CenturyLink) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	cluster := PMXCluster{}
	clc.initProvider()
	clc.login()
	cluster.Cluster = clc.provisionCoreOSCluster(params.ServerCount)
	cluster.Agent = Server{Name: clc.createdDhcpName, PublicIP: clc.publicIP}
	return cluster
}

func (clc *CenturyLink) initProvider() bool {
	clc.groupName = os.Getenv("CLUSTER_NAME")
	clc.apiKey = os.Getenv("API_KEY")
	clc.apiPassword = os.Getenv("API_PASSWORD")
	clc.location = os.Getenv("DATA_CENTER")
	clc.networkName = os.Getenv("NETWORK_NAME")
	clc.serverPassword = fmt.Sprintf("%sA!", clc.randSeq(8))

	if clc.apiKey == "" || clc.apiPassword == "" || clc.groupName == "" || clc.location == "" || clc.networkName == "" {
		panic("\n\nMissing Params...Check Docs....\n\n")
	}

	clc.groupName = fmt.Sprintf("%s-%s", clc.groupName, clc.randSeq(4))
	clc.httpUtil = HttpUtil{APIEndPoint: clc.clcApi}
	return true
}

func (clc *CenturyLink) login() bool {
	var postData = struct {
		APIKEY   string
		Password string
	}{clc.apiKey, clc.apiPassword}

	resp := clc.httpUtil.postJSONData("/auth/logon", postData)

	var status struct {
		success bool
	}

	json.Unmarshal([]byte(resp), &status)
	if !status.success {
		panic("Login Failed, Please check credentials.")
	}
	return true
}

func (clc *CenturyLink) logout() bool {
	println("\nLogging out...")
	clc.httpUtil.HttpClient.Get(clc.clcApi + "/auth/logout")
	return true
}

func (clc *CenturyLink) provisionCoreOSCluster(count int) []Server {

	clc.groupId = clc.createGroup()
	clc.networkName = clc.getNetwork()
	clc.createdDhcpName = clc.createDhcpServer()
	_, publicIP := clc.addPublicIp()

	var coreosServers []Server
	for i := 0; i < clc.serverCount; i++ {
		serverName := clc.createCoreOSServer()
		coreosServers[i] = Server{Name: serverName, PrivateIP: publicIP} //TODO: get IP
		//resizeDisk(coreosServer)
	}

	return coreosServers
}

func (clc *CenturyLink) createGroup() int {
	fmt.Printf("\nCreating Cluster Group in Data Center %s with name: %s", clc.location, clc.groupName)
	var acctLocation = struct {
		Location string
	}{clc.location}

	var parentId int
	var resp = clc.httpUtil.postJSONData("/Group/GetGroups/json", acctLocation)

	var hwGroups struct {
		AccountAlias   string
		HardwareGroups []struct {
			ID            int
			Name          string
			IsSystemGroup bool
		}
	}

	json.Unmarshal([]byte(resp), &hwGroups)

	clc.accountAlias = hwGroups.AccountAlias

	for _, group := range hwGroups.HardwareGroups {
		if strings.Contains(group.Name, clc.location) && group.IsSystemGroup {
			parentId = group.ID
			break
		}
	}

	var postData = struct {
		AccountAlias string
		ParentID     int
		Name         string
		Description  string
	}{clc.accountAlias, parentId, clc.groupName, "CoreOS Cluster"}

	var respNewGroup = clc.httpUtil.postJSONData("/Group/CreateHardwareGroup/json", postData)
	var newGroup struct {
		Group struct {
			ID int
		}
	}

	json.Unmarshal([]byte(respNewGroup), &newGroup)
	return newGroup.Group.ID
}

func (clc *CenturyLink) randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (clc *CenturyLink) getNetwork() string {
	println("\n\nGetting Network Details...")
	var retValue = ""
	var location = struct {
		Location string
	}{clc.location}

	var resp = clc.httpUtil.postJSONData("/Network/GetAccountNetworks/JSON", location)

	var networks struct {
		Networks []struct {
			Name        string
			Description string
		}
	}

	json.Unmarshal([]byte(resp), &networks)

	for _, group := range networks.Networks {
		if strings.Contains(group.Description, clc.networkName) {
			retValue = group.Description
			break
		}
	}
	return retValue
}

func (clc *CenturyLink) deployBlueprintServer(params interface{}) (BlueprintRequestStatus, string) {
	var resp, serverName string
	var reqStatus BlueprintRequestStatus
	resp = clc.httpUtil.postJSONData("/Blueprint/DeployBlueprint/", params)
	serverName = ""
	json.Unmarshal([]byte(resp), &reqStatus)
	if reqStatus.Success {
		for {
			status := clc.getDeploymentStatus(reqStatus.RequestID)
			if status.Success {
				serverName = "Test Me"
				break
			}
			fmt.Print("  .")
			time.Sleep(5000 * time.Millisecond)
		}
	}
	return reqStatus, serverName
}

func (clc *CenturyLink) createDhcpServer() string {
	println("\nCreating DHCP Server")
	params := BlueprintData{
		ID:            1421,
		LocationAlias: clc.location,
		Parameters: []BlueprintParameters{
			{"T3.BuildServerTask.Password", clc.serverPassword},
			{"T3.BuildServerTask.GroupID", strconv.Itoa(clc.groupId)},
			{"T3.BuildServerTask.Network", clc.networkName},
			{"T3.BuildServerTask.PrimaryDNS", "4.4.2.2"},
			{"T3.BuildServerTask.SecondaryDNS", "4.4.2.3"},
			{"79d24724-4335-4c7d-b8ed-2fa59c5e6f97.Alias", clc.dhcpServerAlias},
		}}
	_, serverName := clc.deployBlueprintServer(params)
	return serverName
}

func (clc *CenturyLink) createCoreOSServer() string {
	println("\nCreating CoreOS Server")
	params := BlueprintData{
		ID:            1422,
		LocationAlias: clc.location,
		Parameters: []BlueprintParameters{
			{"TemplateID", "1422"},
			{"T3.BuildServerTask.Password", clc.serverPassword},
			{"T3.BuildServerTask.GroupID", strconv.Itoa(clc.groupId)},
			{"T3.BuildServerTask.Network", clc.networkName},
			{"T3.BuildServerTask.PrimaryDNS", "4.4.2.2"},
			{"T3.BuildServerTask.SecondaryDNS", "4.4.2.3"},
			{"5c69284a-2398-4e5a-8172-57bc44f4a6c9.Alias", clc.coreosServerAlias},
			{"fee73a61-be29-458a-aaa7-41e5e48aec2b.TaskServer", clc.createdDhcpName},
			//{ "fee73a61-be29-458a-aaa7-41e5e48aec2b.T3.CoreOS.SshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDTmIY/i5x5tjmnZLDORIC+/lEmKzGjjj+5S1I1dAQxO923ionVRVepzKhZWlbGa+IfyhoUhCQABXjJQlcbWGbCyDs0m+w2eqB9WwJQRD9zhl+nMv2B173f6EqmxmiRPENQaeXSCLb244xU2zmA7p8h3oPkDD+EonNP/dbfqfkAKq3dQCKNqCRzcMVTgP+0d0UC5I+lhp0mRe6/AhBWwBzdkmHe0N1u5fRgKxIB30mxXMoIv5AkuHuYSyDQRbGPsciL6sa6qhpXMMEpTlBtEO87EqufgoNj1oqYbSGs4qMEjFNVu7CAK8sDbJ+IVvLLbodzqn/mzZ66CfAnJzv797aN cakkineni@Chaitanyas-MacBook-Air.local"},
		}}
	_, serverName := clc.deployBlueprintServer(params)
	return serverName
}

func (clc *CenturyLink) resizeDisk(coreosServer string) {
	fmt.Printf("\nResizing %s", coreosServer)
	params := fmt.Sprintf("{\"AccountAlias\": \"%s\", \"Name\": \"%s\", \"ScsiBusID\": \"0\", \"ScsiDeviceID\": \"2\", \"ResizeGuestDisk\": true, \"NewSizeGB\": 50 }", clc.accountAlias, coreosServer)
	clc.httpUtil.postJSONData("Server/ResizeDisk/json", params)
}

func (clc *CenturyLink) addPublicIp() (bool, string) {
	println("\nAdding Public IP Address....")
	var status bool
	var ipAddress string
	var postData = struct {
		AccountAlias   string
		ServerName     string
		ServerPassword string
		AllowSSH       bool
	}{clc.accountAlias, clc.createdDhcpName, clc.serverPassword, true}

	resp := clc.httpUtil.postJSONData("/Network/AddPublicIPAddress/json", postData)

	var reqStatus BlueprintRequestStatus
	json.Unmarshal([]byte(resp), &reqStatus)

	if !reqStatus.Success {
		fmt.Println("\n%s", resp)
		status = false
		ipAddress = ""
		return status, ipAddress
	}

	for {
		status := clc.getDeploymentStatus(reqStatus.RequestID)
		if status.Success {
			break
		}
		fmt.Print("  .")
		time.Sleep(5000 * time.Millisecond)
	}
	return true, ipAddress
}

func (clc *CenturyLink) getDeploymentStatus(reqId int) BlueprintRequestStatus {
	var postData = struct {
		RequestID     int
		LocationAlias string
	}{reqId, clc.location}

	var reqStatus BlueprintRequestStatus
	resp := clc.httpUtil.postJSONData("/Blueprint/GetBlueprintStatus/json", postData)
	fmt.Printf("\n%s", resp)
	json.Unmarshal([]byte(resp), &reqStatus)
	return reqStatus
}

type BlueprintData struct {
	ID            int
	LocationAlias string
	Parameters    []BlueprintParameters
}

type BlueprintParameters struct {
	Name  string
	Value string
}

type BlueprintRequestStatus struct {
	RequestID int
	Success   bool
}
