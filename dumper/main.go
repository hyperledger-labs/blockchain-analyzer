package main

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"fmt"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"

	"github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricbeatsetup"
	"github.com/hyperledger-elastic/agent/fabricbeat/modules/fabricutils"
	"github.com/hyperledger-elastic/agent/fabricbeat/modules/ledgerutils"
)

func main() {
	fmt.Println("Standalone dumper program started running")

	fbSetup := fabricbeatsetup.FabricbeatSetup{
		OrgName:       "org1",
		ConfigFile:    "/home/prehi/go/src/github.com/hyperledger-elastic/network/basic/connection-profile-1.yaml",
		Peer:          "peer0.org1.el-network.com",
		AdminCertPath: "/home/prehi/go/src/github.com/hyperledger-elastic/network/basic/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/signcerts/Admin@org1.el-network.com-cert.pem",
		AdminKeyPath:  "/home/prehi/go/src/github.com/hyperledger-elastic/network/basic/crypto-config/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp/keystore/adminKey1",
	}

	dumper := &DumperConfig{
		FabricSetup: fbSetup,
		Persistence: DefaultConfig,
	}

	error := fbSetup.Initialize()
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(1)
	}

	lastBlockNums := make(map[*ledger.Client]uint64)

	for _, ledgerClient := range fbSetup.LedgerClients {
		lastBlockNums[ledgerClient] = 0
		blockHeight, err := ledgerutils.GetBlockHeight(ledgerClient)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		for lastBlockNums[ledgerClient] < blockHeight {
			var transactions []string
			block, typeInfo, createdAt, _, err := ledgerutils.ProcessBlock(lastBlockNums[ledgerClient], ledgerClient)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			var channelIdWrapper struct {
				channelId string
			}

			for _, d := range block.Data.Data {
				if typeInfo != "ENDORSER_TRANSACTION" {
					txId, channelId, creator, creatorOrg, _, err := ledgerutils.ProcessTx(d)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}
					channelIdWrapper.channelId = channelId

					dumper.Persistence.PersistNonEndorserTx(
						NonEndorserTx{
							BlockNumber: lastBlockNums[ledgerClient],
							TxID:        txId,
							ChannelID:   channelId,
							CreatedAt:   createdAt,
							Creator:     creator,
							CreatorOrg:  creatorOrg,
							TxType:      typeInfo,
						},
					)
					logp.Info("Non-endorser transaction persisted")

				} else {
					txId, channelId, creator, creatorOrg, txRWSet, chaincodeName, chaincodeVersion, err := ledgerutils.ProcessEndorserTx(d)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
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

								// fmt.Println(fmt.Sprintf("len(bt.Fsetup.Chaincodes) = %d", len(bt.Fsetup.Chaincodes)))

								// //fmt.Println("Chaincode name: " + bt.Fsetup.Chaincodes[chaincodeName].Name + "\n\n\n")

								// //fmt.Println("Linking key: " + bt.Fsetup.Chaincodes[chaincodeName].Linkingkey + "\n\n\n")
								// for _, chaincode := range bt.config.Chaincodes {
								// 	fmt.Println(fmt.Sprintf("Chaincode name: %s, linking key: %s, values length: %d", chaincode.Name, chaincode.Linkingkey, len(chaincode.Values)))
								// }

								// var LinkingkeyString string
								// ccIndex := fabricutils.IndexOfChaincode(bt.Fsetup.Chaincodes, chaincodeName)
								// if ccIndex < 0 || valueMap[bt.config.Chaincodes[ccIndex].Linkingkey] == nil {
								// 	LinkingkeyString = ""
								// } else {
								// 	if str, ok := valueMap[bt.config.Chaincodes[ccIndex].Linkingkey].(string); ok {
								// 		LinkingkeyString = str
								// 	} else {
								// 		return errors.New(fmt.Sprintf("valueMap contains interface{} value instead of string with key %s", bt.config.Chaincodes[ccIndex].Linkingkey))
								// 	}
								// }

								// Sending a new event to the "key" index with the write data

								dumper.Persistence.PersistWrite(
									Write{
										TxID:             txId,
										ChannelID:        channelId,
										ChaincodeName:    chaincodeName,
										ChaincodeVersion: chaincodeVersion,
										Write:            writeset[writeIndex],
										Key:              w.Key,
										//Linkingkey:       string
										Value:      writeset[writeIndex].Value,
										CreatedAt:  createdAt,
										Creator:    creator,
										CreatorOrg: creatorOrg,
									},
								)

								// event := beat.Event{
								// 	Timestamp: time.Now(),
								// 	Fields: libbeatCommon.MapStr{
								// 		"type":              b.Info.Name,
								// 		"tx_id":             txId,
								// 		"channel_id":        channelId,
								// 		"chaincode_name":    chaincodeName,
								// 		"chaincode_version": chaincodeVersion,
								// 		"index_name":        bt.config.KeyIndexName,
								// 		"peer":              bt.config.Peer,
								// 		"write":             writeset[writeIndex],
								// 		"key":               w.Key,
								// 		"linking_key":       LinkingkeyString,
								// 		"value":             writeset[writeIndex].Value,
								// 		"created_at":        createdAt,
								// 		"creator":           creator,
								// 		"creator_org":       creatorOrg,
								// 	},
								// }
								// bt.client.Publish(event)
								logp.Info("Write persisted")
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

					dumper.Persistence.PersistEndorserTx(
						EndorserTx{
							BlockNumber:      lastBlockNums[ledgerClient],
							TxID:             txId,
							ChannelID:        channelId,
							ChaincodeName:    chaincodeName,
							ChaincodeVersion: chaincodeVersion,
							CreatedAt:        createdAt,
							Creator:          creator,
							CreatorOrg:       creatorOrg,
							Readset:          readset,
							Writeset:         writeset,
							TxType:           typeInfo,
						},
					)
					logp.Info("Endorser transaction persisted")
				}
			}
			prevHash := hex.EncodeToString(block.Header.PreviousHash)
			dataHash := hex.EncodeToString(block.Header.DataHash)
			blockHash := fabricutils.GenerateBlockHash(block.Header.PreviousHash, block.Header.DataHash, block.Header.Number)

			dumper.Persistence.PersistBlock(
				Block{
					BlockNumber:  lastBlockNums[ledgerClient],
					ChannelID:    channelIdWrapper.channelId,
					BlockHash:    blockHash,
					PreviousHash: prevHash,
					DataHash:     dataHash,
					CreatedAt:    createdAt,
					transactions: transactions,
				},
			)
			logp.Info("Block persisted")

			lastBlockNums[ledgerClient] += 1
		}
	}
	fmt.Println("Standalone dumper program run successfully")
}
