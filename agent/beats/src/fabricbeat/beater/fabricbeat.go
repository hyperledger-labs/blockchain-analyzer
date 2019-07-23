package beater

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"fabricbeat/config"
	"fabricbeat/modules/ledgerutils"

	"github.com/elastic/beats/libbeat/beat"
	libbeatCommon "github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"

	"github.com/pkg/errors"

	"fabricbeat/modules/elastic"
	"fabricbeat/modules/fabricbeatsetup"
	"fabricbeat/modules/fabricutils"
	"fabricbeat/modules/templates"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done          chan struct{}
	config        config.Config
	client        beat.Client
	Fsetup        *fabricbeatsetup.FabricbeatSetup
	lastBlockNums map[*ledger.Client]uint64
}

// New creates an instance of fabricbeat.
func New(b *beat.Beat, cfg *libbeatCommon.Config) (beat.Beater, error) {
	c := config.DefaultConfig
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
		// Chaincodes:           make(map[string]config.Chaincode),
		Chaincodes: bt.config.Chaincodes,
	}

	// Initializing chaincode data from config
	/*for _, chaincode := range bt.config.Chaincodes {
		fSetup.Chaincodes[chaincode.Name] = chaincode
	}*/

	fmt.Println(fmt.Sprintf("len(fSetup.Chaincodes) = %d", len(fSetup.Chaincodes)))

	// Initialization of the Fabric SDK from the previously set properties
	err1 := fSetup.Initialize()
	if err1 != nil {
		logp.Error(err1)
		return nil, err1
	}
	bt.Fsetup = fSetup

	fmt.Println(fmt.Sprintf("len(fSetup.Chaincodes) = %d", len(fSetup.Chaincodes)))

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

		var err error
		// Get the last known block's number from Elasticsearch.
		lastBlockNumber, err := elastic.GetBlockNumber(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%s/_doc/1", bt.config.Peer, bt.Fsetup.Channels[ledgerClient]))
		if err != nil {
			return err
		}
		// Save last known block number locally for this channel
		bt.lastBlockNums[ledgerClient] = lastBlockNumber.BlockNumber

		logp.Info("Last known block number on channel %s: %d", lastBlockNumber.ChannelId, bt.lastBlockNums[ledgerClient])

		if bt.lastBlockNums[ledgerClient] != 0 {

			// Get block hash of the last known block
			blockHashFromElastic, err := elastic.GetBlockHash(bt.config.ElasticURL, bt.config.BlockIndexName, bt.config.Organization, bt.config.Peer, lastBlockNumber.ChannelId, bt.lastBlockNums[ledgerClient])
			if err != nil {
				return err
			}

			// Retrieve last known block from the ledger
			blockHashFromLedger, err := ledgerutils.GetBlockHash(bt.lastBlockNums[ledgerClient], ledgerClient)
			if err != nil {
				return err
			}
			// Compare block hash from ledger and Elastic
			if blockHashFromLedger != blockHashFromElastic {
				return errors.New(fmt.Sprintf("The hash of the last known block (block number: %d) and the same block on the ledger do not match! Hash from Elastic: %s, hash from ledger: %s", bt.lastBlockNums[ledgerClient], blockHashFromElastic, blockHashFromLedger))
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
		for _, d := range block.Data.Data {
			if typeInfo != "ENDORSER_TRANSACTION" {
				txId, channelId, creator, creatorOrg, _, err := ledgerutils.ProcessTx(d)
				lastBlockNumber.ChannelId = channelId
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
						for writeIndex, w := range ns.KvRwSet.Writes {
							writeset = append(writeset, &fabricutils.Writeset{})
							writeset[writeIndex].Namespace = ns.NameSpace
							writeset[writeIndex].Key = w.Key

							err = json.Unmarshal(w.Value, &writeset[writeIndex].Value)
							if err != nil {
								logp.Warn("Error unmarshaling value into writeset: %s", err.Error())
							}
							// With this map, we can obtain the top level fields of the value.
							var valueMap map[string]interface{}
							err = json.Unmarshal(w.Value, &valueMap)
							if err != nil {
								logp.Warn("Error unmarshaling value into map: %s", err.Error())
							}
							fmt.Println(fmt.Sprintf("\n\n\nSize of map: %d\n\n\n", len(valueMap)))

							writeset[writeIndex].IsDelete = w.IsDelete

							fmt.Println(fmt.Sprintf("len(bt.Fsetup.Chaincodes) = %d", len(bt.Fsetup.Chaincodes)))

							//fmt.Println("Chaincode name: " + bt.Fsetup.Chaincodes[chaincodeName].Name + "\n\n\n")

							//fmt.Println("Linking key: " + bt.Fsetup.Chaincodes[chaincodeName].LinkingKey + "\n\n\n")
							for _, chaincode := range bt.config.Chaincodes {
								fmt.Println(fmt.Sprintf("Chaincode name: %s, linking key: %s, values length: %d", chaincode.Name, chaincode.LinkingKey, len(chaincode.Values)))
							}

							var linkingKeyString string
							ccIndex := fabricutils.IndexOfChaincode(bt.Fsetup.Chaincodes, chaincodeName)
							if ccIndex < 0 || valueMap[bt.config.Chaincodes[ccIndex].LinkingKey] == nil {
								linkingKeyString = ""
							} else {
								if str, ok := valueMap[bt.config.Chaincodes[ccIndex].LinkingKey].(string); ok {
									linkingKeyString = str
								} else {
									return errors.New(fmt.Sprintf("valueMap contains interface{} value instead of string with key %s", bt.config.Chaincodes[ccIndex].LinkingKey))
								}
							}

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
									"write":             writeset[writeIndex],
									"key":               w.Key,
									// "linking_key":       valueMap[bt.Fsetup.Chaincodes[chaincodeName].LinkingKey], // Get the configured linking key name for this chaincode, and use it to obtain linking key from Value
									"linking_key": linkingKeyString,
									"value":       writeset[writeIndex].Value,
									"created_at":  createdAt,
									"creator":     creator,
									"creator_org": creatorOrg,
								},
							}
							bt.client.Publish(event)
							logp.Info("Write event sent")
						}
					}

					if len(ns.KvRwSet.Reads) > 0 {
						// Getting the reads
						for readIndex, w := range ns.KvRwSet.Reads {
							readset = append(readset, &fabricutils.Readset{})
							readset[readIndex].Namespace = ns.NameSpace
							readset[readIndex].Key = w.Key
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

		// TEMP
		type OutmostStruct struct {
			OuterStruct struct {
				MiddleStruct struct {
					InnerStruct struct {
						Flag string
					}
				}
			}
		}

		var testStruct OutmostStruct
		var testEmptyInterface interface{}
		testStruct.OuterStruct.MiddleStruct.InnerStruct.Flag = "FLAG"
		testStructBytes, err := json.Marshal(testStruct)
		if err != nil {
			return err
		}
		err = json.Unmarshal(testStructBytes, &testEmptyInterface)

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: libbeatCommon.MapStr{
				"type":          b.Info.Name,
				"block_number":  lastBlockNumber.BlockNumber,
				"channel_id":    lastBlockNumber.ChannelId,
				"block_hash":    blockHash,
				"previous_hash": prevHash,
				"data_hash":     dataHash,
				"test_struct":   testEmptyInterface,
				"created_at":    createdAt,
				"index_name":    bt.config.BlockIndexName,
				"peer":          bt.config.Peer,
				"transactions":  transactions,
			},
		}
		bt.client.Publish(event)
		logp.Info("Block event sent")

		// Send the latest known block number to Elasticsearch
		err = elastic.SendBlockNumber(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%s/_doc/1", bt.config.Peer, bt.Fsetup.Channels[ledgerClient]), lastBlockNumber)
		if err != nil {
			return err
		}
		bt.lastBlockNums[ledgerClient]++
		lastBlockNumber.BlockNumber++
	}
	return nil
}
