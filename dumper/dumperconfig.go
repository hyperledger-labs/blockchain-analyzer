package main

import (
	"time"

	"github.com/blockchain-analyzer/agent/agentmodules/fabricbeatsetup"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
)

// Defines the setup and the persistence interface, keeps track of the last known blocks for each channel
type DumperConfig struct {
	Period        time.Duration
	FabricSetup   *fabricbeatsetup.FabricbeatSetup
	LastBlockNums map[*ledger.Client]uint64
	Persistence   Persistent
}
