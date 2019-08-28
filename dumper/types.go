package main

import (
	"time"

	"github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricutils"
)

// Struct for storing non-endorser (most likely config) transaction data
type NonEndorserTx struct {
	BlockNumber uint64
	ChannelID   string
	CreatedAt   time.Time
	Creator     string
	CreatorOrg  string
	TxType      string
}

// Struct for storing endorser transaction data
type EndorserTx struct {
	BlockNumber      uint64
	TxID             string
	ChannelID        string
	ChaincodeName    string
	ChaincodeVersion string
	CreatedAt        time.Time
	Creator          string
	CreatorOrg       string
	TxType           string
	Readset          []*fabricutils.Readset
	Writeset         []*fabricutils.Writeset
}

// Struct for storing the data of one key write
type Write struct {
	TxID             string
	ChannelID        string
	ChaincodeName    string
	ChaincodeVersion string
	Write            *fabricutils.Writeset
	Key              string
	Linkingkey       string
	Value            interface{}
	CreatedAt        time.Time
	Creator          string
	CreatorOrg       string
}

// Struct for storing Block data
type Block struct {
	BlockNumber  uint64
	ChannelID    string
	BlockHash    string
	PreviousHash string
	DataHash     string
	CreatedAt    time.Time
	transactions []string
}
