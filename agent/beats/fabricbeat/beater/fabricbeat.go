package beater

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/balazsprehoda/fabricbeat/modules/ledgerutils"

	"github.com/elastic/beats/libbeat/beat"
	libbeatCommon "github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	fabricbeatConfig "github.com/balazsprehoda/fabricbeat/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"

	"github.com/pkg/errors"

	"github.com/balazsprehoda/fabricbeat/modules/elastic"
	"github.com/balazsprehoda/fabricbeat/modules/fabricbeatsetup"
	"github.com/balazsprehoda/fabricbeat/modules/fabricutils"
	"github.com/balazsprehoda/fabricbeat/modules/templates"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done          chan struct{}
	config        fabricbeatConfig.Config
	client        beat.Client
	Fsetup        *fabricbeatsetup.FabricbeatSetup
	lastBlockNums map[*ledger.Client]uint64
}

// New creates an instance of fabricbeat.
func New(b *beat.Beat, cfg *libbeatCommon.Config) (beat.Beater, error) {
	c := fabricbeatConfig.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Fabricbeat{
		done:          make(chan struct{}),
		config:        c,
		lastBlockNums: make(map[*ledger.Client]uint64),
	}

	fSetup := &fabricbeatsetup.FabricbeatSetup{
		OrgName:              bt.config.Organization,
		ConfigFile:           bt.config.ConnectionProfile,
		Peer:                 bt.config.Peer,
		AdminCertPath:        bt.config.AdminCertPath,
		AdminKeyPath:         bt.config.AdminKeyPath,
		ElasticURL:           bt.config.ElasticURL,
		KibanaURL:            bt.config.KibanaURL,
		BlockIndexName:       bt.config.BlockIndexName,
		TransactionIndexName: bt.config.TransactionIndexName,
		KeyIndexName:         bt.config.KeyIndexName,
		DashboardDirectory:   bt.config.DashboardDirectory,
		TemplateDirectory:    bt.config.TemplateDirectory,
	}

	// Initialization of the Fabric SDK from the previously set properties
	err1 := fSetup.Initialize()
	if err1 != nil {
		logp.Error(err1)
		return nil, err1
	}
	bt.Fsetup = fSetup

	// Generate the index patterns and dashboards for the connected peer from templates in the kibana_templates folder
	err := templates.GenerateDashboards(fSetup)
	if err != nil {
		return nil, err
	}

	return bt, nil
}

// Run starts fabricbeat.
func (bt *Fabricbeat) Run(b *beat.Beat) error {
	logp.Info("fabricbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	// Ramp-up section
	// Iterate over the known channels (one ledger client per channel)
	err = bt.rampUp(b)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		logp.Info("Start event loop")
		// Iterate over the known channels (one ledger client per channel)
		for index, ledgerClient := range bt.Fsetup.LedgerClients {
			var lastBlockNumber elastic.BlockNumber
			lastBlockNumber.BlockNumber = bt.lastBlockNums[ledgerClient]
			err = bt.ProcessNewBlocks(b, ledgerClient, index)
			if err != nil {
				return err
			}
		}

	}
}

// Stop stops fabricbeat.
func (bt *Fabricbeat) Stop() {
	bt.client.Close()
	close(bt.done)
	defer bt.Fsetup.CloseSDK()
}

// Helps the Fabricbeat agent to continue where it left off: Gets the last known block from Elasticsearch, compares it to the ledger,
// and if the two match, it queries every block since the last known block, and sends their data to Elasticsearch. If the block retrieved from Elasticsearch
// and the block from the ledger do not match, it returns an error.
func (bt *Fabricbeat) rampUp(b *beat.Beat) error {
	for index, ledgerClient := range bt.Fsetup.LedgerClients {

		var lastBlockNumber elastic.BlockNumber
		var err error
		// Get the last known block's number from Elasticsearch.
		lastBlockNumber.BlockNumber, err = elastic.GetBlockNumber(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index))
		if err != nil {
			return err
		}
		// Save last known block number locally for this channel
		bt.lastBlockNums[ledgerClient] = lastBlockNumber.BlockNumber

		logp.Info("Last known block number: %d", bt.lastBlockNums[ledgerClient])

		if bt.lastBlockNums[ledgerClient] != 0 {

			// Get block hash of the last known block
			blockHashFromElastic, err := elastic.GetBlockHash(bt.config.ElasticURL, bt.config.BlockIndexName, bt.config.Organization, bt.config.Peer, bt.lastBlockNums[ledgerClient])
			if err != nil {
				return err
			}

			// Retrieve last known block from the ledger
			blockHash, err := ledgerutils.GetBlockHash(bt.lastBlockNums[ledgerClient], ledgerClient)
			if err != nil {
				return err
			}
			// Compare block hash from ledger and Elastic
			if blockHash != blockHashFromElastic {
				return errors.New(fmt.Sprintf("The hash of the last known block (block number: %d) and the same block on the ledger do not match!", bt.lastBlockNums[ledgerClient]))
			} else {
				logp.Info(fmt.Sprintf("The hash of the last known block (block number: %d) and the same block on the ledger match.", bt.lastBlockNums[ledgerClient]))
			}
			// Increase last block number, so that the querying starts from the next block
			bt.lastBlockNums[ledgerClient]++
		}
		err = bt.ProcessNewBlocks(b, ledgerClient, index)
		if err != nil {
			return err
		}
	}
	return nil
}

// Gets the blocks from the ledger and sends their data to Elasticsearch.
func (bt *Fabricbeat) ProcessNewBlocks(b *beat.Beat, ledgerClient *ledger.Client, index int) error {

	var lastBlockNumber elastic.BlockNumber
	lastBlockNumber.BlockNumber = bt.lastBlockNums[ledgerClient]

	blockHeight, err := ledgerutils.GetBlockHeight(ledgerClient)
	if err != nil {
		return err
	}
	for lastBlockNumber.BlockNumber < blockHeight {
		var transactions []string
		block, typeInfo, createdAt, _, err := ledgerutils.ProcessBlock(lastBlockNumber.BlockNumber, ledgerClient)
		if err != nil {
			return err
		}
		for i, d := range block.Data.Data {
			if typeInfo != "ENDORSER_TRANSACTION" {
				txId, channelId, creator, creatorOrg, _, err := ledgerutils.ProcessTx(d)
				if err != nil {
					return err
				}
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: libbeatCommon.MapStr{
						"type":             b.Info.Name,
						"block_number":     lastBlockNumber.BlockNumber,
						"tx_id":            txId,
						"channel_id":       channelId,
						"index_name":       bt.config.TransactionIndexName,
						"peer":             bt.config.Peer,
						"created_at":       createdAt,
						"creator":          creator,
						"creator_org":      creatorOrg,
						"transaction_type": typeInfo,
					},
				}
				bt.client.Publish(event)
				logp.Info("Non-endorser transaction event sent")

			} else {
				txId, channelId, creator, creatorOrg, txRWSet, chaincodeName, chaincodeVersion, err := ledgerutils.ProcessEndorserTx(d)
				if err != nil {
					return err
				}
				readset := []*fabricutils.Readset{}
				writeset := []*fabricutils.Writeset{}
				// Getting read-write set
				// For every namespace
				for _, ns := range txRWSet.NsRwSets {

					if len(ns.KvRwSet.Writes) > 0 {
						// Getting the writes
						for _, w := range ns.KvRwSet.Writes {
							writeset = append(writeset, &fabricutils.Writeset{})
							writeset[i].Namespace = ns.NameSpace
							writeset[i].Key = w.Key
							err = json.Unmarshal(w.Value, &writeset[i].Value)

							// if err != nil {
							// 	return err
							// }
							//writeset[i].Value = string(w.Value)
							writeset[i].IsDelete = w.IsDelete

							// Sending a new event to the "key" index with the write data
							event := beat.Event{
								Timestamp: time.Now(),
								Fields: libbeatCommon.MapStr{
									"type":              b.Info.Name,
									"tx_id":             txId,
									"channel_id":        channelId,
									"chaincode_name":    chaincodeName,
									"chaincode_version": chaincodeVersion,
									"index_name":        bt.config.KeyIndexName,
									"peer":              bt.config.Peer,
									"key":               w.Key,
									"previous_key":      writeset[i].Value["previousKey"],
									"value":             string(w.Value),
									"created_at":        createdAt,
									"creator":           creator,
									"creator_org":       creatorOrg,
								},
							}
							bt.client.Publish(event)
							logp.Info("Write event sent")
						}
					}

					if len(ns.KvRwSet.Reads) > 0 {
						// Getting the reads
						for i, w := range ns.KvRwSet.Reads {
							readset = append(readset, &fabricutils.Readset{})
							readset[i].Namespace = ns.NameSpace
							readset[i].Key = w.Key
						}
					}
				}

				transactions = append(transactions, txId)
				// Sending the transaction data to the "transaction" index
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: libbeatCommon.MapStr{
						"type":              b.Info.Name,
						"block_number":      lastBlockNumber.BlockNumber,
						"tx_id":             txId,
						"channel_id":        channelId,
						"chaincode_name":    chaincodeName,
						"chaincode_version": chaincodeVersion,
						"index_name":        bt.config.TransactionIndexName,
						"peer":              bt.config.Peer,
						"created_at":        createdAt,
						"creator":           creator,
						"creator_org":       creatorOrg,
						"readset":           readset,
						"writeset":          writeset,
						"transaction_type":  typeInfo,
					},
				}
				bt.client.Publish(event)
				logp.Info("Endorsement transaction event sent")
			}
		}
		prevHash := hex.EncodeToString(block.Header.PreviousHash)
		dataHash := hex.EncodeToString(block.Header.DataHash)
		blockHash := fabricutils.GenerateBlockHash(block.Header.PreviousHash, block.Header.DataHash, block.Header.Number)
		// Sending the block data to the "block" index
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: libbeatCommon.MapStr{
				"type":          b.Info.Name,
				"block_number":  lastBlockNumber.BlockNumber,
				"block_hash":    blockHash,
				"previous_hash": prevHash,
				"data_hash":     dataHash,
				"created_at":    createdAt,
				"index_name":    bt.config.BlockIndexName,
				"peer":          bt.config.Peer,
				"transactions":  transactions,
			},
		}
		bt.client.Publish(event)
		logp.Info("Block event sent")

		// Send the latest known block number to Elasticsearch
		err = elastic.SendBlockNumber(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index), lastBlockNumber)
		if err != nil {
			return err
		}
		bt.lastBlockNums[ledgerClient]++
		lastBlockNumber.BlockNumber++
	}
	return nil
}
