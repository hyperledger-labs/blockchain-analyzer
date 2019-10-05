package main

import (
	"time"

	"github.com/blockchain-analyzer/agent/agentmodules/fabricsetup"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
)

// Defines the setup and the persistence interface, keeps track of the last known blocks for each channel
type DumperConfig struct {
	Period        time.Duration
	FabricSetup   *fabricsetup.FabricSetup
	LastBlockNums map[*ledger.Client]uint64
	Persistence   Persistent
}
