package beater

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	//"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/util" /*fabricCommon*/
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
		OrgName:       bt.config.Organization,
		ConfigFile:    bt.config.ConnectionProfile,
		Peer:          bt.config.Peer,
		AdminCertPath: bt.config.AdminCertPath,
		AdminKeyPath:  bt.config.AdminKeyPath,
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

	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		logp.Info("Start event loop")
		for _, ledgerClient := range bt.fSetup.ledgerClients {
			infoResponse, err1 := ledgerClient.QueryInfo()
			if err1 != nil {
				logp.Warn("QueryInfo returned error: " + err1.Error())
			}
			queriedBlockNum := infoResponse.BCI.Height
			logp.Info("Block height: %d", queriedBlockNum)
			// bt.lastBlockNums[ledgerClient] = 6
			for bt.lastBlockNums[ledgerClient] < queriedBlockNum {
				blockResponse, blockError := ledgerClient.QueryBlock(bt.lastBlockNums[ledgerClient])
				if blockError != nil {
					logp.Warn("QueryBlock returned error: " + blockError.Error())
				}
				bt.lastBlockNums[ledgerClient]++

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
				createdAt := time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))
				fmt.Printf("--------------------- Timestamp: %s", createdAt)

				fmt.Printf("There are %d transactions in this block\n", len(blockResponse.Data.Data))

				// Checking validity
				txsFltr := util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])

				// Processing transactions
				for i, d := range blockResponse.Data.Data {
					fmt.Printf("tx %d (validation status: %s):\n", i, txsFltr.Flag(i).String())

					env, err := utils.GetEnvelopeFromBlock(d)
					if err != nil {
						fmt.Printf("Error getting tx from block(%s)", err)
						os.Exit(-1)
					}

					// GetPayload can only handle endorsement transactions
					payload, err := utils.GetPayload(env)
					if err != nil {
						fmt.Printf("GetPayload returns err %s", err)
						os.Exit(-1)
					}
					chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
					if err != nil {
						fmt.Printf("UnmarshalChannelHeader returns err %s", err)
						os.Exit(-1)
					}

					shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
					if err != nil {
						fmt.Printf("GetSignatureHeader returns err %s", err)
						os.Exit(-1)
					}

					tx, err := utils.GetTransaction(payload.Data)
					if err != nil {
						fmt.Printf("GetTransaction returns err %s", err)
						os.Exit(-1)
					}

					_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
					if payloadErr != nil {
						fmt.Printf("GetPayloads returns err %s This is not an endorsement transaction, it must be config!", payloadErr)
						event := beat.Event{
							Timestamp: time.Now(),
							Fields: libbeatCommon.MapStr{
								"type":             b.Info.Name,
								"block_number":     blockResponse.Header.Number,
								"tx_id":            chdr.TxId,
								"channel_id":       chdr.ChannelId,
								"index_name":       "transaction",
								"created_at":       createdAt,
								"creator":          returnCreatorString(shdr.Creator),
								"transaction_type": typeInfo,
							},
						}
						bt.client.Publish(event)
						logp.Info("Config transaction event sent")
					}

					fmt.Printf("\tCH: %s\n", chdr.ChannelId)
					fmt.Printf("\tcreator: %s\n", returnCreatorString(shdr.Creator))
					if payloadErr == nil {
						fmt.Printf("\tCC: %+v\n", respPayload.ChaincodeId)

						txRWSet := &rwsetutil.TxRwSet{}
						err = txRWSet.FromProtoBytes(respPayload.Results)
						if err != nil {
							fmt.Printf("FromProtoBytes returns err %s", err)
							os.Exit(-1)
						}

						readset := []*Readset{}
						writeset := []*Writeset{}
						fmt.Printf("\tRead-Write set:\n")
						readIndex := 0
						writeIndex := 0
						for _, ns := range txRWSet.NsRwSets {
							fmt.Printf("\t\tNamespace: %s\n", ns.NameSpace)

							if len(ns.KvRwSet.Writes) > 0 {
								fmt.Printf("\t\t\tWrites:\n")
								for _, w := range ns.KvRwSet.Writes {
									fmt.Printf("\t\t\t\tK: %s, V:%s\n", w.Key, strings.Replace(string(w.Value), "\n", ".", -1))
									fmt.Printf("Write as string: %s", w.String())
									writeset = append(writeset, &Writeset{})
									writeset[writeIndex].Namespace = ns.NameSpace
									writeset[writeIndex].Key = w.Key
									writeset[writeIndex].Value = string(w.Value)
									writeset[writeIndex].IsDelete = w.IsDelete
									writeIndex++
								}
							}

							if len(ns.KvRwSet.Reads) > 0 {
								fmt.Printf("\t\t\tReads:\n")
								for _, w := range ns.KvRwSet.Reads {
									fmt.Printf("\t\t\t\tK: %s\n", w.Key)
									readset = append(readset, &Readset{})
									readset[readIndex].Namespace = ns.NameSpace
									readset[readIndex].Key = w.Key
									readIndex++
								}
							}

							if len(ns.CollHashedRwSets) > 0 {
								for _, c := range ns.CollHashedRwSets {
									fmt.Printf("\t\t\tCollection: %s\n", c.CollectionName)

									if len(c.HashedRwSet.HashedWrites) > 0 {
										fmt.Printf("\t\t\t\tWrites:\n")
										for _, ww := range c.HashedRwSet.HashedWrites {
											fmt.Printf("\t\t\t\t\tK: %s, V:%s\n",
												base64.StdEncoding.EncodeToString(ww.KeyHash),
												base64.StdEncoding.EncodeToString(ww.ValueHash))
										}
									}

									if len(c.HashedRwSet.HashedReads) > 0 {
										fmt.Printf("\t\t\t\tReads:\n")
										for _, ww := range c.HashedRwSet.HashedReads {
											fmt.Printf("\t\t\t\t\tK: %s\n",
												base64.StdEncoding.EncodeToString(ww.KeyHash))
										}
									}
								}
							}
						}
						transactions = append(transactions, chdr.TxId)
						event := beat.Event{
							Timestamp: time.Now(),
							Fields: libbeatCommon.MapStr{
								"type":              b.Info.Name,
								"block_number":      blockResponse.Header.Number,
								"tx_id":             chdr.TxId,
								"channel_id":        chdr.ChannelId,
								"chaincode_name":    respPayload.ChaincodeId.Name,
								"chaincode_version": respPayload.ChaincodeId.Version,
								"index_name":        "transaction",
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
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: libbeatCommon.MapStr{
						"type":          b.Info.Name,
						"block_number":  blockResponse.Header.Number,
						"block_hash":    blockHash,
						"previous_hash": prevHash,
						"data_hash":     dataHash,
						"created_at":    createdAt,
						"index_name":    "block",
						"transactions":  transactions,
					},
				}
				bt.client.Publish(event)
				logp.Info("Block event sent")
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
	ConfigFile    string
	initialized   bool
	OrgName       string
	Peer          string
	mspClient     *msp.Client
	AdminCertPath string
	AdminKeyPath  string
	adminIdentity providerMSP.SigningIdentity
	resClient     *resmgmt.Client
	ledgerClients []*ledger.Client
	sdk           *fabsdk.FabricSDK
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

	//prevHash := strToHex(previousHashStr)

	//dataHash := strToHex(dataHashStr)

	h := common.BlockHeader{
		Number:       blockNumber,
		PreviousHash: previousHash,
		DataHash:     dataHash}

	return hex.EncodeToString(h.Hash())
}

// Closes SDK
func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
