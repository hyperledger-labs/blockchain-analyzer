package main

import (
	"time"

	"github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricutils"
)

type NonEndorserTx struct {
	BlockNumber uint64
	TxID        string
	ChannelID   string
	CreatedAt   time.Time
	Creator     string
	CreatorOrg  string
	TxType      string
}

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

type Block struct {
	BlockNumber  uint64
	ChannelID    string
	BlockHash    string
	PreviousHash string
	DataHash     string
	CreatedAt    time.Time
	transactions []string
}
