package main

import "github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricbeatsetup"

type DumperConfig struct {
	FabricSetup fabricbeatsetup.FabricbeatSetup
	Persistence Persistent
}
