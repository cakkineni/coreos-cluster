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
	letters           []rune
}

func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	cl.clcApi = "https://api.tier3.com/rest"
	cl.dhcpServerAlias = "DHCP"
	cl.coreosServerAlias = "COREOS"
	cl.letters = []rune("abcdefghijklmnopqrstuvwxyz")
	return cl
}

func (clc CenturyLink) ProvisionPMXCluster(params ClusterParams) PMXCluster {
	println("\nProvisioning PMX Cluster in Centurylink")
	cluster := PMXCluster{}
	clc.initProvider()
	clc.login()

	clc.groupId = clc.createGroup()
	println("GroupId:", clc.groupId)
	clc.networkName = clc.getNetwork()
	println("NetworkId:", clc.networkName)
	cluster.Agent = clc.createDhcpServer()
	_, publicIP := clc.addPublicIp(cluster.Agent)
	println("PublicIP:", publicIP)
	cluster.Agent.PublicIP = publicIP
	println("Deploying clusters of coreos...")
	cluster.Cluster = clc.provisionCoreOSCluster(params.ServerCount, cluster.Agent.Name)
	return cluster
}

func (clc *CenturyLink) initProvider() bool {
	clc.apiKey = os.Getenv("API_KEY")
	clc.apiPassword = os.Getenv("API_PASSWORD")
	clc.location = os.Getenv("REGION")
	clc.networkName = os.Getenv("NETWORK_NAME")
	clc.serverPassword = fmt.Sprintf("%sA!", clc.randSeq(8))

	if clc.apiKey == "" || clc.apiPassword == "" || clc.location == "" || clc.networkName == "" {
		println("\n\nMissing Params...Check Docs....\n\n")
	}

	clc.groupName = fmt.Sprintf("coreos-%s", clc.randSeq(4))
	clc.httpUtil = NewHttpUtil() //HttpUtil{APIEndPoint: clc.clcApi}
	clc.httpUtil.APIEndPoint = clc.clcApi
	return true
}

func (clc *CenturyLink) login() bool {
	var postData = struct {
		APIKEY   string
		Password string
	}{clc.apiKey, clc.apiPassword}

	resp := clc.httpUtil.postJSONData("/auth/logon", postData)

	println(resp)

	var status struct {
		Success bool
	}

	json.Unmarshal([]byte(resp), &status)
	if !status.Success {
		panic("Login Failed, Please check credentials.")
	}
	return true
}

func (clc *CenturyLink) logout() bool {
	println("\nLogging out...")
	clc.httpUtil.HttpClient.Get(clc.clcApi + "/auth/logout")
	return true
}

func (clc *CenturyLink) provisionCoreOSCluster(count int, dhcpServerName string) []Server {
	println("00000000")
	var coreosServers []Server
	for i := 0; i < count; i++ {
		println("Provisioning Server ", i+1)
		server := clc.createCoreOSServer(dhcpServerName)
		coreosServers = append(coreosServers, server)
		//resizeDisk(coreosServer)
	}

	return coreosServers
}

func (clc *CenturyLink) createGroup() int {
	var acctLocation = struct {
		Location string
	}{clc.location}

	var parentId int
	var resp = clc.httpUtil.postJSONData("/Group/GetGroups/json", acctLocation)

	println(resp)

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
		b[i] = clc.letters[rand.Intn(len(clc.letters))]
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

	println(resp)

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

func (clc *CenturyLink) deployBlueprintServer(params interface{}) Server {
	var resp, serverName string
	var reqStatus BlueprintRequestStatus
	resp = clc.httpUtil.postJSONData("/Blueprint/DeployBlueprint/", params)

	println(resp)
	serverName = ""
	json.Unmarshal([]byte(resp), &reqStatus)
	if reqStatus.Success {
		for {
			status := clc.getDeploymentStatus(reqStatus.RequestID)
			if strings.Contains(strings.ToLower(status.CurrentStatus), "succeeded") {
				serverName = status.Servers[0]
				break
			}
			fmt.Print("  .")
			time.Sleep(5000 * time.Millisecond)
		}
	}

	if serverName != "" {
		var serverData = struct {
			Name         string
			AccountAlias string
		}{serverName, clc.accountAlias}
		resp = clc.httpUtil.postJSONData("/Server/GetServer/", serverData)
		println(resp)
		var clcServer struct {
			Server struct {
				IPAddresses []struct {
					Address     string
					AddressType string
				}
			}
		}
		json.Unmarshal([]byte(resp), &clcServer)
		return Server{Name: serverName, PrivateIP: clcServer.Server.IPAddresses[0].Address}
	}
	return Server{Name: serverName}
}

func (clc *CenturyLink) createDhcpServer() Server {
	println("\nCreating DHCP Server")
	params := BlueprintData{
		ID:            1196,
		LocationAlias: clc.location,
		Parameters: []BlueprintParameters{
			{"T3.BuildServerTask.Password", clc.serverPassword},
			{"T3.BuildServerTask.GroupID", strconv.Itoa(clc.groupId)},
			{"T3.BuildServerTask.Network", clc.networkName},
			{"T3.BuildServerTask.PrimaryDNS", "4.4.2.2"},
			{"T3.BuildServerTask.SecondaryDNS", "4.4.2.3"},
			{"79d24724-4335-4c7d-b8ed-2fa59c5e6f97.Alias", clc.dhcpServerAlias},
		}}
	server := clc.deployBlueprintServer(params)
	return server
}

func (clc *CenturyLink) createCoreOSServer(dhcpServerName string) Server {
	println("\nCreating CoreOS Server")
	params := BlueprintData{
		ID:            1197,
		LocationAlias: clc.location,
		Parameters: []BlueprintParameters{
			{"TemplateID", "1422"},
			{"T3.BuildServerTask.Password", clc.serverPassword},
			{"T3.BuildServerTask.GroupID", strconv.Itoa(clc.groupId)},
			{"T3.BuildServerTask.Network", clc.networkName},
			{"T3.BuildServerTask.PrimaryDNS", "4.4.2.2"},
			{"T3.BuildServerTask.SecondaryDNS", "4.4.2.3"},
			{"5c69284a-2398-4e5a-8172-57bc44f4a6c9.Alias", clc.coreosServerAlias},
			{"fee73a61-be29-458a-aaa7-41e5e48aec2b.TaskServer", dhcpServerName},
			//{ "fee73a61-be29-458a-aaa7-41e5e48aec2b.T3.CoreOS.SshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDTmIY/i5x5tjmnZLDORIC+/lEmKzGjjj+5S1I1dAQxO923ionVRVepzKhZWlbGa+IfyhoUhCQABXjJQlcbWGbCyDs0m+w2eqB9WwJQRD9zhl+nMv2B173f6EqmxmiRPENQaeXSCLb244xU2zmA7p8h3oPkDD+EonNP/dbfqfkAKq3dQCKNqCRzcMVTgP+0d0UC5I+lhp0mRe6/AhBWwBzdkmHe0N1u5fRgKxIB30mxXMoIv5AkuHuYSyDQRbGPsciL6sa6qhpXMMEpTlBtEO87EqufgoNj1oqYbSGs4qMEjFNVu7CAK8sDbJ+IVvLLbodzqn/mzZ66CfAnJzv797aN cakkineni@Chaitanyas-MacBook-Air.local"},
		}}
	server := clc.deployBlueprintServer(params)
	return server
}

func (clc *CenturyLink) resizeDisk(coreosServer string) {
	fmt.Printf("\nResizing %s", coreosServer)
	params := fmt.Sprintf("{\"AccountAlias\": \"%s\", \"Name\": \"%s\", \"ScsiBusID\": \"0\", \"ScsiDeviceID\": \"2\", \"ResizeGuestDisk\": true, \"NewSizeGB\": 50 }", clc.accountAlias, coreosServer)
	clc.httpUtil.postJSONData("Server/ResizeDisk/json", params)
}

func (clc *CenturyLink) addPublicIp(server Server) (bool, string) {
	println("\nAdding Public IP Address....")
	var status bool
	var ipAddress string
	var postData = struct {
		AccountAlias string
		ServerName   string
		IPAddress    string
		AllowSSH     bool
		Location     string
	}{clc.accountAlias, server.Name, server.PrivateIP, true, clc.location}

	resp := clc.httpUtil.postJSONData("/Network/AddPublicIPAddress/json", postData)
	println(resp)

	var reqStatus BlueprintRequestStatus
	json.Unmarshal([]byte(resp), &reqStatus)

	if !reqStatus.Success {
		status = false
		ipAddress = ""
		return status, ipAddress
	}

	for {
		status := clc.getDeploymentStatus(reqStatus.RequestID)
		if status.CurrentStatus == "Succeeded" {
			var serverData = struct {
				Name         string
				AccountAlias string
			}{server.Name, clc.accountAlias}
			resp = clc.httpUtil.postJSONData("/Server/GetServer/", serverData)
			var clcServerInfo struct {
				Server struct {
					IPAddresses []struct {
						Address     string
						AddressType string
					}
				}
			}
			json.Unmarshal([]byte(resp), &clcServerInfo)
			println(resp)
			ipAddress = clcServerInfo.Server.IPAddresses[1].Address
			break
		}
		fmt.Print("  .")
		time.Sleep(5000 * time.Millisecond)
	}
	println(ipAddress)
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
	RequestID     int
	Success       bool
	CurrentStatus string
	Servers       []string
}
