package fabricbeatsetup

import (
	"github.com/blockchain-analyzer/agent/agentmodules/fabricsetup"
)

// Fabric, Elasticsearch and Kibana specific setup
type FabricbeatSetup struct {
	initialized          bool
	OrgName              string
	ElasticURL           string
	KibanaURL            string
	DashboardDirectory   string
	TemplateDirectory    string
	BlockIndexName       string
	TransactionIndexName string
	KeyIndexName         string
	Chaincodes           []fabricsetup.Chaincode
}

