// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period               time.Duration `config:"period"`
	Organization         string        `config:"organization"`
	Peer                 string        `config:"peer"`
	ConnectionProfile    string        `config:"connectionProfile"`
	AdminCertPath        string        `config:"adminCertPath"`
	AdminKeyPath         string        `config:"adminKeyPath"`
	ElasticURL           string        `config:"elasticURL"`
	KibanaURL            string        `config:"kibanaURL"`
	BlockIndexName       string        `config:"blockIndexName"`
	TransactionIndexName string        `config:"transactionIndexName"`
	KeyIndexName         string        `config:"keyIndexName"`
	DashboardDirectory   string        `config:"dashboardDirectory"`
	TemplateDirectory    string        `config:"templateDirectory"`
	Chaincodes           []Chaincode   `config:"chaincodes"`
}

type Chaincode struct {
	Name       string   //`chaincode:"name"`
	Linkingkey string   //`chaincode:"linkingKey"`
	Values     []string //`chaincode:"values"`
}

var DefaultConfig = Config{
	Period:               1 * time.Second,
	Organization:         "org1",
	Peer:                 "peer0.org1.el-network.com",
	ConnectionProfile:    "connection.yaml",
	AdminCertPath:        "/home/prehi/internship/testNetwork/hyperledger-elastic/network/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/signcerts/Admin@org1.el-network.com-cert.pem",
	AdminKeyPath:         "/home/prehi/internship/testNetwork/hyperledger-elastic/network/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/keystore/adminKey1",
	ElasticURL:           "http://localhost:9200",
	KibanaURL:            "http://localhost:5601",
	BlockIndexName:       "block",
	TransactionIndexName: "transaction",
	KeyIndexName:         "key",
	DashboardDirectory:   "/home/prehi/internship/testNetwork/hyperledger-elastic/dashboards",
	TemplateDirectory:    "/home/prehi/internship/testNetwork/hyperledger-elastic/agent/kibana_templates",
	Chaincodes: []Chaincode{
		Chaincode{
			Name:       "mycc",
			Linkingkey: "linking_key",
			Values:     []string{"myvalue"},
		},
	},
}
