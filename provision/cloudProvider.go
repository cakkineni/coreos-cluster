package provision

import "strings"

type CloudProvider interface {
	ProvisionPMXCluster(params ClusterParams) PMXCluster
}

func New(providerType string) CloudProvider {
	providerType = strings.ToLower(providerType)
	switch providerType {
	case "amz":
		return Amazon{}.New()
	case "clc":
		return CenturyLink{}.New()
	case "docean":
		return DOcean{}.New()
	}
	return nil
}

type ClusterParams struct {
	ServerCount        int
	CloudConfigCluster string
	CloudConfigAgent   string
}

type PMXCluster struct {
	Agent   *Server
	Cluster []Server
}

type Server struct {
	PublicIP  string
	PrivateIP string
	Name      string
}
