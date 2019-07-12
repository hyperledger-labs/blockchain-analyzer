package beater

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	libbeatCommon "github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/gogo/protobuf/proto"

	fabricbeatConfig "github.com/balazsprehoda/fabricbeat/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	providerMSP "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	protoMSP "github.com/hyperledger/fabric/protos/msp"

	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done          chan struct{}
	config        fabricbeatConfig.Config
	client        beat.Client
	fSetup        *FabricSetup
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

	fSetup := &FabricSetup{
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

	bt.fSetup = fSetup

	// Get all channels the peer is part of
	channelsResponse, err2 := fSetup.resClient.QueryChannels(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err2 != nil {
		return nil, err2
	}

	// Initialize the ledger client for each channel
	for _, channel := range channelsResponse.Channels {
		logp.Info(channel.ChannelId)
		channelContext := fSetup.sdk.ChannelContext(channel.ChannelId, fabsdk.WithIdentity(fSetup.adminIdentity))
		if channelContext == nil {
			logp.Warn("Channel context creation failed, ChannelContext() returned nil for channel " + channel.ChannelId)
		}
		ledgerClient, err4 := ledger.New(channelContext)
		if err4 != nil {
			return nil, err4
		}
		fSetup.ledgerClients = append(fSetup.ledgerClients, ledgerClient)
		logp.Info("Ledger client initialized for channel " + channel.ChannelId)
	}
	logp.Info("Channel clients initialized")

	// Get installed chaincodes of the peer
	chaincodeResponse, err3 := fSetup.resClient.QueryInstalledChaincodes(resmgmt.WithTargetEndpoints(fSetup.Peer))
	if err3 != nil {
		return nil, err3
	}
	for _, chaincode := range chaincodeResponse.Chaincodes {
		logp.Info(chaincode.Name)
	}

	// Generate the index patterns and dashboards for the connected peer from templates in the kibana_templates folder
	err := fSetup.GenerateDashboards()
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
	for index, ledgerClient := range bt.fSetup.ledgerClients {

		var lastBlockNumber BlockNumber
		// Get the last known block number from Elasticsearch
		resp, err := http.Get(fmt.Sprintf(bt.fSetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index))
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
			lastBlockNumber := lastBlockNumberResponse.BlockNumber
			// Save last known block number for this channel
			bt.lastBlockNums[ledgerClient] = lastBlockNumber.BlockNumber
		}

		logp.Info("Last known block number: %d", bt.lastBlockNums[ledgerClient])

		if bt.lastBlockNums[ledgerClient] != 0 {

			// Get the index from which we want to get the last known block
			resp, err = http.Get(fmt.Sprintf(bt.config.ElasticURL+"/_cat/indices/fabricbeat-*%s*%s*", bt.config.BlockIndexName, bt.config.Peer))
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
			requestBody := fmt.Sprintf("{\"query\": {\"term\": {\"block_number\": {\"value\":%d}}}}", bt.lastBlockNums[ledgerClient])
			logp.Debug("URL for last block query: %s", url)
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
			blockResponse, blockError := ledgerClient.QueryBlock(bt.lastBlockNums[ledgerClient])
			if blockError != nil {
				logp.Warn("QueryBlock returned error: " + blockError.Error())
			}
			logp.Info("Querying last known block from ledger successful")

			blockHash := generateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
			if blockHash != blockHashFromElastic {
				return errors.New(fmt.Sprintf("The hash of the last known block (block number: %d) and the same block on the ledger do not match!", blockResponse.Header.Number))
			} else {
				logp.Info(fmt.Sprintf("The hash of the last known block (block number: %d) and the same block on the ledger match.", blockResponse.Header.Number))
				// Increase last block number, so that the querying starts from the next block
				bt.lastBlockNums[ledgerClient]++
			}
		}

		// Get the block height of this channel
		infoResponse, err1 := ledgerClient.QueryInfo()
		if err1 != nil {
			logp.Warn("QueryInfo returned error: " + err1.Error())
		}

		// Query all blocks since the last known
		for bt.lastBlockNums[ledgerClient] < infoResponse.BCI.Height {
			blockResponse, blockError := ledgerClient.QueryBlock(bt.lastBlockNums[ledgerClient])
			if blockError != nil {
				logp.Warn("QueryBlock returned error: " + blockError.Error())
			}

			// Transaction Ids in this block
			var transactions []string

			// Getting block creation timestamp
			env, err := utils.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
			if err != nil {
				fmt.Printf("GetEnvelopeFromBlock returned error: %s", err)
			}

			channelHeader, err := utils.ChannelHeader(env)
			if err != nil {
				fmt.Printf("ChannelHeader returned error: %s", err)
			}
			typeCode := channelHeader.GetType()
			typeInfo := TypeCodeToInfo(typeCode)
			createdAt := time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))

			// Checking validity
			txsFltr := util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])

			// Processing transactions
			for i, d := range blockResponse.Data.Data {
				fmt.Printf("tx %d (validation status: %s):\n", i, txsFltr.Flag(i).String())

				env, err := utils.GetEnvelopeFromBlock(d)
				if err != nil {
					return err
				}

				payload, err := utils.GetPayload(env)
				if err != nil {
					return err
				}
				chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
				if err != nil {
					return err
				}

				shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
				if err != nil {
					return err
				}

				tx, err := utils.GetTransaction(payload.Data)
				if err != nil {
					return err
				}

				if typeInfo != "ENDORSER_TRANSACTION" {
					event := beat.Event{
						Timestamp: time.Now(),
						Fields: libbeatCommon.MapStr{
							"type":             b.Info.Name,
							"block_number":     blockResponse.Header.Number,
							"tx_id":            chdr.TxId,
							"channel_id":       chdr.ChannelId,
							"index_name":       bt.config.TransactionIndexName,
							"created_at":       createdAt,
							"creator":          returnCreatorString(shdr.Creator),
							"transaction_type": typeInfo,
						},
					}
					bt.client.Publish(event)
					logp.Info("Config transaction event sent")
				} else {
					_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
					if payloadErr != nil {
						return payloadErr
					}

					txRWSet := &rwsetutil.TxRwSet{}
					err = txRWSet.FromProtoBytes(respPayload.Results)
					if err != nil {
						return err
					}

					readset := []*Readset{}
					writeset := []*Writeset{}
					// Getting read-write set
					readIndex := 0
					writeIndex := 0
					// For every namespace
					for _, ns := range txRWSet.NsRwSets {

						if len(ns.KvRwSet.Writes) > 0 {
							// Getting the writes
							for _, w := range ns.KvRwSet.Writes {
								writeset = append(writeset, &Writeset{})
								writeset[writeIndex].Namespace = ns.NameSpace
								writeset[writeIndex].Key = w.Key
								writeset[writeIndex].Value = string(w.Value)
								writeset[writeIndex].IsDelete = w.IsDelete
								writeIndex++

								// Sending a new event to the "key" index with the write data
								event := beat.Event{
									Timestamp: time.Now(),
									Fields: libbeatCommon.MapStr{
										"type":              b.Info.Name,
										"tx_id":             chdr.TxId,
										"channel_id":        chdr.ChannelId,
										"chaincode_name":    respPayload.ChaincodeId.Name,
										"chaincode_version": respPayload.ChaincodeId.Version,
										"index_name":        bt.config.KeyIndexName,
										"key":               w.Key,
										"value":             string(w.Value),
										"created_at":        createdAt,
										"creator":           returnCreatorString(shdr.Creator),
									},
								}
								bt.client.Publish(event)
								logp.Info("Write event sent")
							}
						}

						if len(ns.KvRwSet.Reads) > 0 {
							// Getting the reads
							for _, w := range ns.KvRwSet.Reads {
								readset = append(readset, &Readset{})
								readset[readIndex].Namespace = ns.NameSpace
								readset[readIndex].Key = w.Key
								readIndex++
							}
						}
					}
					transactions = append(transactions, chdr.TxId)
					// Sending the transaction data to the "transaction" index
					event := beat.Event{
						Timestamp: time.Now(),
						Fields: libbeatCommon.MapStr{
							"type":              b.Info.Name,
							"block_number":      blockResponse.Header.Number,
							"tx_id":             chdr.TxId,
							"channel_id":        chdr.ChannelId,
							"chaincode_name":    respPayload.ChaincodeId.Name,
							"chaincode_version": respPayload.ChaincodeId.Version,
							"index_name":        bt.config.TransactionIndexName,
							"created_at":        createdAt,
							"creator":           returnCreatorString(shdr.Creator),
							"readset":           readset,
							"writeset":          writeset,
							"transaction_type":  typeInfo,
						},
					}
					bt.client.Publish(event)
					logp.Info("Endorsement transaction event sent")
				}

			}

			prevHash := hex.EncodeToString(blockResponse.Header.PreviousHash)
			dataHash := hex.EncodeToString(blockResponse.Header.DataHash)
			blockHash := generateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
			// Sending the block data to the "block" index
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: libbeatCommon.MapStr{
					"type":          b.Info.Name,
					"block_number":  blockResponse.Header.Number,
					"block_hash":    blockHash,
					"previous_hash": prevHash,
					"data_hash":     dataHash,
					"created_at":    createdAt,
					"index_name":    bt.config.BlockIndexName,
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
			resp, err = http.Post(fmt.Sprintf(bt.fSetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index), "application/json", bytes.NewBuffer(jsonBlockNumber))
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
		for index, ledgerClient := range bt.fSetup.ledgerClients {

			// Get block height for this channel
			infoResponse, err1 := ledgerClient.QueryInfo()
			if err1 != nil {
				logp.Warn("QueryInfo returned error: " + err1.Error())
			}

			// Query all blocks since the last known
			for bt.lastBlockNums[ledgerClient] < infoResponse.BCI.Height {
				blockResponse, blockError := ledgerClient.QueryBlock(bt.lastBlockNums[ledgerClient])
				if blockError != nil {
					logp.Warn("QueryBlock returned error: " + blockError.Error())
				}

				// Transaction Ids in this block
				var transactions []string

				// Getting block creation timestamp
				env, err := utils.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
				if err != nil {
					fmt.Printf("GetEnvelopeFromBlock returned error: %s", err)
				}

				channelHeader, err := utils.ChannelHeader(env)
				if err != nil {
					fmt.Printf("ChannelHeader returned error: %s", err)
				}
				typeCode := channelHeader.GetType()
				typeInfo := TypeCodeToInfo(typeCode)

				createdAt := time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))

				// Checking validity
				txsFltr := util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])

				// Processing transactions
				for i, d := range blockResponse.Data.Data {
					fmt.Printf("tx %d (validation status: %s):\n", i, txsFltr.Flag(i).String())

					env, err := utils.GetEnvelopeFromBlock(d)
					if err != nil {
						return err
					}

					payload, err := utils.GetPayload(env)
					if err != nil {
						return err
					}
					chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
					if err != nil {
						return err
					}

					shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
					if err != nil {
						return err
					}

					tx, err := utils.GetTransaction(payload.Data)
					if err != nil {
						return err
					}

					if typeInfo != "ENDORSER_TRANSACTION" {
						event := beat.Event{
							Timestamp: time.Now(),
							Fields: libbeatCommon.MapStr{
								"type":             b.Info.Name,
								"block_number":     blockResponse.Header.Number,
								"tx_id":            chdr.TxId,
								"channel_id":       chdr.ChannelId,
								"index_name":       bt.config.TransactionIndexName,
								"created_at":       createdAt,
								"creator":          returnCreatorString(shdr.Creator),
								"transaction_type": typeInfo,
							},
						}
						bt.client.Publish(event)
						logp.Info("Config transaction event sent")
					} else {
						_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
						if payloadErr != nil {
							return payloadErr
						}

						txRWSet := &rwsetutil.TxRwSet{}
						err = txRWSet.FromProtoBytes(respPayload.Results)
						if err != nil {
							return err
						}

						readset := []*Readset{}
						writeset := []*Writeset{}
						// Getting read-write set
						readIndex := 0
						writeIndex := 0
						// For every namespace
						for _, ns := range txRWSet.NsRwSets {

							if len(ns.KvRwSet.Writes) > 0 {
								// Getting the writes
								for _, w := range ns.KvRwSet.Writes {
									writeset = append(writeset, &Writeset{})
									writeset[writeIndex].Namespace = ns.NameSpace
									writeset[writeIndex].Key = w.Key
									writeset[writeIndex].Value = string(w.Value)
									writeset[writeIndex].IsDelete = w.IsDelete
									writeIndex++

									// Sending a new event to the "key" index with the write data
									event := beat.Event{
										Timestamp: time.Now(),
										Fields: libbeatCommon.MapStr{
											"type":              b.Info.Name,
											"tx_id":             chdr.TxId,
											"channel_id":        chdr.ChannelId,
											"chaincode_name":    respPayload.ChaincodeId.Name,
											"chaincode_version": respPayload.ChaincodeId.Version,
											"index_name":        bt.config.KeyIndexName,
											"key":               w.Key,
											"value":             string(w.Value),
											"created_at":        createdAt,
											"creator":           returnCreatorString(shdr.Creator),
										},
									}
									bt.client.Publish(event)
									logp.Info("Write event sent")
								}
							}

							if len(ns.KvRwSet.Reads) > 0 {
								// Getting the reads
								for _, w := range ns.KvRwSet.Reads {
									readset = append(readset, &Readset{})
									readset[readIndex].Namespace = ns.NameSpace
									readset[readIndex].Key = w.Key
									readIndex++
								}
							}
						}
						transactions = append(transactions, chdr.TxId)
						// Sending the transaction data to the "transaction" index
						event := beat.Event{
							Timestamp: time.Now(),
							Fields: libbeatCommon.MapStr{
								"type":              b.Info.Name,
								"block_number":      blockResponse.Header.Number,
								"tx_id":             chdr.TxId,
								"channel_id":        chdr.ChannelId,
								"chaincode_name":    respPayload.ChaincodeId.Name,
								"chaincode_version": respPayload.ChaincodeId.Version,
								"index_name":        bt.config.TransactionIndexName,
								"created_at":        createdAt,
								"creator":           returnCreatorString(shdr.Creator),
								"readset":           readset,
								"writeset":          writeset,
								"transaction_type":  typeInfo,
							},
						}
						bt.client.Publish(event)
						logp.Info("Endorsement transaction event sent")
					}

				}

				prevHash := hex.EncodeToString(blockResponse.Header.PreviousHash)
				dataHash := hex.EncodeToString(blockResponse.Header.DataHash)
				blockHash := generateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
				// Sending the block data to the "block" index
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: libbeatCommon.MapStr{
						"type":          b.Info.Name,
						"block_number":  blockResponse.Header.Number,
						"block_hash":    blockHash,
						"previous_hash": prevHash,
						"data_hash":     dataHash,
						"created_at":    createdAt,
						"index_name":    bt.config.BlockIndexName,
						"transactions":  transactions,
					},
				}
				bt.client.Publish(event)
				logp.Info("Block event sent")

				// Send the latest known block number to Elasticsearch
				var lastBlockNumber BlockNumber
				lastBlockNumber.BlockNumber = bt.lastBlockNums[ledgerClient]
				jsonBlockNumber, err := json.Marshal(lastBlockNumber)
				if err != nil {
					return err
				}
				resp, err := http.Post(fmt.Sprintf(bt.fSetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index), "application/json", bytes.NewBuffer(jsonBlockNumber))
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 && resp.StatusCode != 201 {
					return errors.New("Sending last block number to Elasticsearch failed!")
				}
				bt.lastBlockNums[ledgerClient]++
			}
		}

	}
}

// Stop stops fabricbeat.
func (bt *Fabricbeat) Stop() {
	bt.client.Close()
	close(bt.done)
	defer bt.fSetup.CloseSDK()
}

// FabricSetup implementation
type FabricSetup struct {
	ConfigFile           string
	initialized          bool
	OrgName              string
	Peer                 string
	ElasticURL           string
	KibanaURL            string
	mspClient            *msp.Client
	AdminCertPath        string
	AdminKeyPath         string
	adminIdentity        providerMSP.SigningIdentity
	resClient            *resmgmt.Client
	ledgerClients        []*ledger.Client
	sdk                  *fabsdk.FabricSDK
	BlockIndexName       string
	TransactionIndexName string
	KeyIndexName         string
	DashboardDirectory   string
	TemplateDirectory    string
}

// Initialize reads the configuration file and sets up FabricSetup
func (setup *FabricSetup) Initialize() error {

	logp.Info("Initializing SDK")
	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("SDK already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err0 := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err0 != nil {
		logp.Warn("SDK initialization failed!")
		return errors.WithMessage(err0, "failed to create SDK")
	}
	setup.sdk = sdk
	logp.Info("SDK created")

	mspContext := setup.sdk.Context(fabsdk.WithOrg(setup.OrgName))
	if mspContext == nil {
		logp.Warn("setup.sdk.Context() returned nil")
	}

	cert, err := ioutil.ReadFile(setup.AdminCertPath)
	if err != nil {
		return err
	}
	pk, err := ioutil.ReadFile(setup.AdminKeyPath)
	if err != nil {
		return err
	}

	var err1 error
	setup.mspClient, err1 = msp.New(mspContext)
	if err1 != nil {
		return err1
	}

	id, err2 := setup.mspClient.CreateSigningIdentity(providerMSP.WithCert(cert), providerMSP.WithPrivateKey(pk))
	if err2 != nil {
		return err2
	}

	setup.adminIdentity = id
	resContext := setup.sdk.Context(fabsdk.WithIdentity(id))
	var err7 error
	setup.resClient, err7 = resmgmt.New(resContext)
	if err7 != nil {
		return errors.WithMessage(err7, "failed to create new resmgmt client")
	}
	logp.Info("Resmgmt client created")

	logp.Info("Initialization Successful")
	setup.initialized = true
	return nil
}

// Generates the index patterns and dashboards for the connected peer from templates in the kibana_templates folder.
func (setup *FabricSetup) GenerateDashboards() error {

	// The beginnings of the dashboard template names (i.e. overview-dashboard-TEMPLATE.json -> overview)
	dashboardNames := []string{"overview", "block", "key", "transaction"}
	visualizationNames := []string{"block_count", "transaction_count", "transaction_per_organization", "transaction_count_timeline"}
	templates := []string{"block", "transaction", "key"}
	var patternId string
	// Create index patterns for the peer the agent connects to
	for _, templateName := range templates {
		// Load index pattern template and replace title
		logp.Info("Creating %s index pattern for connected peer", templateName)

		indexPatternJSON, err := ioutil.ReadFile(fmt.Sprintf("%s/%s-index-pattern-TEMPLATE.json", setup.TemplateDirectory, templateName))
		if err != nil {
			return err
		}

		indexPatternJSONstring := string(indexPatternJSON)
		// Replace id placeholders (in URL formatted fields)
		for _, dashboardName := range dashboardNames {

			// Replace dashboard id placeholders
			idExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re := regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-dashboard-%s-%s", dashboardName, setup.Peer, setup.OrgName))

			// Replace search id placeholders
			idExpression = fmt.Sprintf("%s_SEARCH_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re = regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-search-%s-%s", dashboardName, setup.Peer, setup.OrgName))

			// Replace visualization id placeholders
			idExpression = fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re = regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-visualization-%s-%s", dashboardName, setup.Peer, setup.OrgName))
		}

		// Replace title placeholders
		titleExpression := fmt.Sprintf("INDEX_PATTERN_TEMPLATE_TITLE")
		re := regexp.MustCompile(titleExpression)
		indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("fabricbeat*%s*%s", templateName, setup.Peer))

		indexPatternJSON = []byte(indexPatternJSONstring)
		// Send index pattern to Kibana via Kibana Saved Objects API
		logp.Info("Persisting %s index pattern for connected peer", templateName)

		patternId = fmt.Sprintf("fabricbeat-%s-%s", templateName, setup.Peer)
		request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/saved_objects/index-pattern/%s", setup.KibanaURL, patternId), bytes.NewBuffer(indexPatternJSON))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("kbn-xsrf", "true")
		httpClient := http.Client{}
		resp, err := httpClient.Do(request)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// Check if the index pattern exists. 409 is for version conflict, it means that this index pattern has already been created.
		// TODO check for existing index pattern and replace it.
		if resp.StatusCode != 200 && resp.StatusCode != 409 {
			return errors.New(fmt.Sprintf("Failed to create %s index pattern:\nResponse status code: %d\nResponse body: %s", templateName, resp.StatusCode, string(body)))
		}
		logp.Info("%s index pattern created", templateName)
	}

	for _, dashboardName := range dashboardNames {

		// Load dashboard template
		logp.Info("Creating %s dashboard from template", dashboardName)
		dashboardBytes, err := ioutil.ReadFile(fmt.Sprintf("/home/prehi/go/src/github.com/balazsprehoda/fabricbeat/kibana_templates/%s-dashboard-TEMPLATE.json", dashboardName))
		if err != nil {
			return err
		}
		dashboard := string(dashboardBytes)

		for _, templateName := range templates {
			// Replace index pattern id placeholders
			patternId = fmt.Sprintf("fabricbeat-%s-%s", templateName, setup.Peer)
			patternExpression := fmt.Sprintf("%s_PATTERN", strings.ToUpper(templateName))
			re := regexp.MustCompile(patternExpression)
			dashboard = re.ReplaceAllString(string(dashboard), patternId)

			// Replace search id placeholders
			searchId := fmt.Sprintf("%s-search-%s-%s", templateName, setup.Peer, setup.OrgName)
			searchIdExpression := fmt.Sprintf("%s_SEARCH_TEMPLATE_ID", strings.ToUpper(templateName))
			re = regexp.MustCompile(searchIdExpression)
			dashboard = re.ReplaceAllString(dashboard, searchId)

			// Replace search title placeholders
			searchTitle := fmt.Sprintf("%s Search %s (%s)", strings.Title(templateName), setup.Peer, setup.OrgName)
			searchTitleExpression := fmt.Sprintf("%s_SEARCH_TEMPLATE_TITLE", strings.ToUpper(templateName))
			re = regexp.MustCompile(searchTitleExpression)
			dashboard = re.ReplaceAllString(dashboard, searchTitle)
		}

		for _, visualizationName := range visualizationNames {
			// Replace visualization id placeholders
			visualizationId := fmt.Sprintf("%s-visualization-%s-%s", visualizationName, setup.Peer, setup.OrgName)
			visualizationIdExpression := fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_ID", strings.ToUpper(visualizationName))
			re := regexp.MustCompile(visualizationIdExpression)
			dashboard = re.ReplaceAllString(dashboard, visualizationId)

			// Replace visualization title placeholders
			visualizationTitle := fmt.Sprintf("%s Visualization %s (%s)", strings.Title(visualizationName), setup.Peer, setup.OrgName)
			visualizationTitleExpression := fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_TITLE", strings.ToUpper(visualizationName))
			re = regexp.MustCompile(visualizationTitleExpression)
			dashboard = re.ReplaceAllString(dashboard, visualizationTitle)
		}

		// Replace dashboard id
		idExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_ID", strings.ToUpper(dashboardName))
		re := regexp.MustCompile(idExpression)
		dashboard = re.ReplaceAllString(string(dashboard), fmt.Sprintf("%s-dashboard-%s-%s", dashboardName, setup.Peer, setup.OrgName))

		// Replace dashboard title
		titleExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_TITLE", strings.ToUpper(dashboardName))
		re = regexp.MustCompile(titleExpression)
		dashboard = re.ReplaceAllString(string(dashboard), fmt.Sprintf("%s Dashboard %s (%s)", strings.Title(dashboardName), setup.Peer, setup.OrgName))

		// Persist the created dashboard in the configured directory, from where it is going to be loaded
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s-%s-%s.json", setup.DashboardDirectory, dashboardName, setup.Peer, setup.OrgName), []byte(dashboard), 0664)
		if err != nil {
			return err
		}
	}

	return nil
}

// This function is borrowed from an opensource project: https://github.com/ale-aso/fabric-examples/blob/master/blockparser.go
func returnCreatorString(bytes []byte) string {
	defaultString := strings.Replace(string(bytes), "\n", ".", -1)

	sId := &protoMSP.SerializedIdentity{}
	err := proto.Unmarshal(bytes, sId)
	if err != nil {
		return defaultString
	}

	bl, _ := pem.Decode(sId.IdBytes)
	if bl == nil {
		return defaultString
	}

	return sId.Mspid
}

type Readset struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type Writeset struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsDelete  bool   `json:"isDelete"`
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

// This function is borrowed from an opensource project: https://github.com/denisglotov/fabric-hash-calculator
func strToHex(str string) []byte {
	str = strings.TrimPrefix(str, "0x")
	strBytes := []byte(str)
	dst := make([]byte, hex.DecodedLen(len(strBytes)))
	n, err := hex.Decode(dst, strBytes)
	if err != nil {
		log.Fatal(err)
	}
	return dst[:n]
}

func generateBlockHash(previousHash, dataHash []byte, blockNumber uint64) string {

	h := common.BlockHeader{
		Number:       blockNumber,
		PreviousHash: previousHash,
		DataHash:     dataHash}

	return hex.EncodeToString(h.Hash())
}

// Decodes the type of the transaction into string
func TypeCodeToInfo(typeCode int32) string {
	var typeInfo string
	switch typeCode {
	case 0:
		typeInfo = "MESSAGE"
	case 1:
		typeInfo = "CONFIG"
	case 2:
		typeInfo = "CONFIG_UPDATE"
	case 3:
		typeInfo = "ENDORSER_TRANSACTION"
	case 4:
		typeInfo = "ORDERER_TRANSACTION"
	case 5:
		typeInfo = "DELIVER_SEEK_INFO"
	case 6:
		typeInfo = "CHAINCODE_PACKAGE"
	case 8:
		typeInfo = "PEER_ADMIN_OPERATION"
	case 9:
		typeInfo = "TOKEN_TRANSACTION"
	default:
		typeInfo = "UNRECOGNIZED_TYPE"
	}
	return typeInfo
}

// Closes SDK
func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
