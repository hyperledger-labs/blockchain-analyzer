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
	BlockIndexName       string        `config:"blockIndexName"`
	TransactionIndexName string        `config:"transactionIndexName"`
	KeyIndexName         string        `config:"keyIndexName"`
}

var DefaultConfig = Config{
	Period:               1 * time.Second,
	Organization:         "org1",
	Peer:                 "peer0.org1.el-network.com",
	ConnectionProfile:    "connection.yaml",
	AdminCertPath:        "/home/prehi/internship/testNetwork/hyperledger-elastic/network/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/signcerts/Admin@org1.el-network.com-cert.pem",
	AdminKeyPath:         "/home/prehi/internship/testNetwork/hyperledger-elastic/network/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/keystore/adminKey1",
	ElasticURL:           "http://localhost:9200",
	BlockIndexName:       "block",
	TransactionIndexName: "transaction",
	KeyIndexName:         "key",
}
