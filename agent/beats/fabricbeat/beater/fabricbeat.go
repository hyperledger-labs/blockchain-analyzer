package beater

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/balazsprehoda/fabricbeat/modules/ledgerutils"

	"github.com/elastic/beats/libbeat/beat"
	libbeatCommon "github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	fabricbeatConfig "github.com/balazsprehoda/fabricbeat/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/pkg/errors"

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

	// Get all channels the peer is part of
	channelsResponse, err2 := fSetup.ResClient.QueryChannels(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err2 != nil {
		return nil, err2
	}

	// Initialize the ledger client for each channel
	for _, channel := range channelsResponse.Channels {
		channelContext := fSetup.SDK.ChannelContext(channel.ChannelId, fabsdk.WithIdentity(fSetup.AdminIdentity))
		if channelContext == nil {
			logp.Warn("Channel context creation failed, ChannelContext() returned nil for channel " + channel.ChannelId)
		}
		ledgerClient, err4 := ledger.New(channelContext)
		if err4 != nil {
			return nil, err4
		}
		fSetup.LedgerClients = append(fSetup.LedgerClients, ledgerClient)
		logp.Info("Ledger client initialized for channel " + channel.ChannelId)
	}
	logp.Info("Channel clients initialized")

	// Get installed chaincodes of the peer
	chaincodeResponse, err3 := fSetup.ResClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err3 != nil {
		return nil, err3
	}
	for _, chaincode := range chaincodeResponse.Chaincodes {
		logp.Info(chaincode.Name)
	}

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
			var lastBlockNumber BlockNumber
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

// This struct is for parsing block filter query response from Elasticsearch
type BlockIndexFilterResponse struct {
	BlockIndexFilterHitsObject struct {
		BlockIndexFilterHit []struct {
			BlockIndexData struct {
				BlockHash string `json:"block_hash"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type BlockNumber struct {
	BlockNumber uint64 `json:"blockNumber"`
}

type BlockNumberResponse struct {
	Index       string      `json:"_index"`
	Type        string      `json:"_type"`
	Id          string      `json:"_id"`
	BlockNumber BlockNumber `json:"_source"`
}

type SearchRequest struct {
	Size        int `json:"size"`
	QueryObject struct {
		Bool struct {
			FilterObject []struct {
				Term struct {
				}
			} `json:"filter"`
		} `json:"bool"`
	} `json:"query"`
}

// Helps the Fabricbeat agent to continue where it left off: Gets the last known block from Elasticsearch, compares it to the ledger,
// and if the two match, it queries every block since the last known block, and sends their data to Elasticsearch. If the block retrieved from Elasticsearch
// and the block from the ledger do not match, it returns an error.
func (bt *Fabricbeat) rampUp(b *beat.Beat) error {
	for index, ledgerClient := range bt.Fsetup.LedgerClients {

		var lastBlockNumber BlockNumber
		// Get the last known block number from Elasticsearch
		resp, err := http.Get(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 && resp.StatusCode != 404 {
			return errors.New(fmt.Sprintf("Failed getting the last block number! Http response status code: %d", resp.StatusCode))
		}
		if resp.StatusCode == 404 {
			// It is the very first start of the agent, there is no last block yet.
			bt.lastBlockNums[ledgerClient] = 0
			lastBlockNumber.BlockNumber = 0
			logp.Info("Last known block number not found, setting to 0")
		} else {
			// Get the block number info from the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var lastBlockNumberResponse BlockNumberResponse
			err = json.Unmarshal(body, &lastBlockNumberResponse)
			if err != nil {
				return err
			}
			lastBlockNumber = lastBlockNumberResponse.BlockNumber
			// Save last known block number locally for this channel
			bt.lastBlockNums[ledgerClient] = lastBlockNumber.BlockNumber
		}

		logp.Info("Last known block number: %d", bt.lastBlockNums[ledgerClient])

		if bt.lastBlockNums[ledgerClient] != 0 {

			// Get the index from which we want to get the last known block
			resp, err = http.Get(fmt.Sprintf(bt.config.ElasticURL+"/_cat/indices/fabricbeat-*%s*%s*", bt.config.BlockIndexName, bt.config.Organization))
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			// [2] is the name of the index
			blockIndex := strings.Fields(string(body))[2]

			// Retrieve the last known block from Elasticsearch
			httpClient := &http.Client{}
			url := fmt.Sprintf("%s/%s/_search", bt.config.ElasticURL, blockIndex)
			//requestBody := fmt.Sprintf("{\"query\": {\"term\": {\"block_number\": {\"value\":%d}}}", bt.lastBlockNums[ledgerClient])
			requestBody := fmt.Sprintf(`{
				"size": 1,
				"query": {
				  "bool": {
					"filter": [
					  {
						"term": {
						  "block_number": {
							"value": "%d"
						  }
						}
					  },
					  {
						"term": {
						  "peer": {
							"value": "%s"
						  }
						}
					  }
					]
				  }
				},
				"sort": [
				  {
					"value": {
					  "order": "desc"
					}
				  }
				]
			}`, bt.lastBlockNums[ledgerClient], bt.config.Peer)
			logp.Debug("URL for last block query: ", url)
			request, err := http.NewRequest("GET", url, bytes.NewBufferString(requestBody))
			if err != nil {
				return err
			}
			request.Header.Add("Content-Type", "application/json")
			resp, err = httpClient.Do(request)
			if err != nil {
				return err
			}
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return errors.New("Failed to get last block from Elasticsearch: " + string(body))
			}
			var lastBlockResponseFromElastic BlockIndexFilterResponse
			err = json.Unmarshal(body, &lastBlockResponseFromElastic)
			if err != nil {
				return err
			}
			fmt.Println(string(body))
			if lastBlockResponseFromElastic.BlockIndexFilterHitsObject.BlockIndexFilterHit == nil {
				return errors.New("Could not properly unmarshal the response body to BlockIndexFilterResponse: BlockIndexFilterResponse.BlockIndexFilterHitsObject.BlockIndexFilterHit is nil")
			}

			blockHashFromElastic := lastBlockResponseFromElastic.BlockIndexFilterHitsObject.BlockIndexFilterHit[0].BlockIndexData.BlockHash

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

func (bt *Fabricbeat) ProcessNewBlocks(b *beat.Beat, ledgerClient *ledger.Client, index int) error {

	var lastBlockNumber BlockNumber
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
				txId, channelId, creator, _, err := ledgerutils.ProcessTx(d)
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
						"transaction_type": typeInfo,
					},
				}
				bt.client.Publish(event)
				logp.Info("Non-endorser transaction event sent")

			} else {
				txId, channelId, creator, txRWSet, chaincodeName, chaincodeVersion, err := ledgerutils.ProcessEndorserTx(d)
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
							writeset[i].Value = string(w.Value)
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
									"value":             string(w.Value),
									"created_at":        createdAt,
									"creator":           creator,
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
		jsonBlockNumber, err := json.Marshal(lastBlockNumber)
		if err != nil {
			return err
		}
		resp, err := http.Post(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index), "application/json", bytes.NewBuffer(jsonBlockNumber))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return errors.New("Sending last block number to Elasticsearch failed!")
		}
		bt.lastBlockNums[ledgerClient]++
		lastBlockNumber.BlockNumber++
	}
	return nil
}
