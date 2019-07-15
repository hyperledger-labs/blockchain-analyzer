package ledgerutils

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	protoCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"

	"github.com/balazsprehoda/fabricbeat/modules/fabricutils"
	"github.com/elastic/beats/libbeat/logp"
)

func GetBlockHash(blockNumber uint64, ledgerClient *ledger.Client) (string, error) {
	blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
	if blockError != nil {
		return "", blockError
	}
	logp.Info("Querying last known block from ledger successful")

	blockHash := fabricutils.GenerateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
	return blockHash, nil
}

func GetBlockHeight(ledgerClient *ledger.Client) (uint64, error) {
	// Get the block height of this channel
	infoResponse, err := ledgerClient.QueryInfo()
	if err != nil {
		return 0, err
	}
	blockHeight := infoResponse.BCI.Height
	return blockHeight, nil
}

func ProcessBlock(blockNumber uint64, ledgerClient *ledger.Client) (blockResponse *protoCommon.Block, typeInfo string, createdAt time.Time, txsFltr util.TxValidationFlags, err error) {
	blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
	if blockError != nil {
		return nil, "", time.Now(), nil, blockError
	}

	// Getting block creation timestamp
	env, err := utils.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
	if err != nil {
		return nil, "", time.Now(), nil, err
	}

	channelHeader, err := utils.ChannelHeader(env)
	if err != nil {
		return nil, "", time.Now(), nil, err
	}
	typeCode := channelHeader.GetType()
	typeInfo = fabricutils.TypeCodeToInfo(typeCode)
	createdAt = time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))

	// Checking validity
	txsFltr = util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	return
}

func ProcessTx(txData []byte) (txId, channelId, creator string, tx *peer.Transaction, err error) {
	env, err := utils.GetEnvelopeFromBlock(txData)
	if err != nil {
		return "", "", "", nil, err
	}

	payload, err := utils.GetPayload(env)
	if err != nil {
		return "", "", "", nil, err
	}
	chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return "", "", "", nil, err
	}

	shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return "", "", "", nil, err
	}

	tx, err = utils.GetTransaction(payload.Data)
	if err != nil {
		return "", "", "", nil, err
	}

	return chdr.TxId, chdr.ChannelId, fabricutils.ReturnCreatorString(shdr.Creator), tx, nil
}

func ProcessEndorserTx(txData []byte) (txId, channelId, creator string, txRWSet *rwsetutil.TxRwSet, chaincodeName, chaincodeVersion string, err error) {

	txId, channelId, creator, tx, err := ProcessTx(txData)
	if err != nil {
		return "", "", "", nil, "", "", err
	}

	_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
	if payloadErr != nil {
		return "", "", "", nil, "", "", err
	}

	txRWSet = &rwsetutil.TxRwSet{}
	err = txRWSet.FromProtoBytes(respPayload.Results)
	if err != nil {
		return "", "", "", nil, "", "", err
	}

	return txId, channelId, creator, txRWSet, respPayload.ChaincodeId.Name, respPayload.ChaincodeId.Version, nil
}

// // Queries all the blocks since the blockNumber from the ledger, and sends their data to Elasticsearch. Returns the last block number, or error.
// func CatchUp(blockNumber uint64, lastBlockNumber fabricutils.BlockNumber, ledgerClient *ledger.Client, b *beat.Beat, beatClient beat.Client) (uint64, error) {
// 	// Get the block height of this channel
// 	infoResponse, err1 := ledgerClient.QueryInfo()
// 	if err1 != nil {
// 		logp.Warn("QueryInfo returned error: " + err1.Error())
// 	}

// 	// Query all blocks since the last known
// 	for blockNumber < infoResponse.BCI.Height {
// 		blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
// 		if blockError != nil {
// 			logp.Warn("QueryBlock returned error: " + blockError.Error())
// 		}

// 		// Transaction Ids in this block
// 		var transactions []string

// 		// Getting block creation timestamp
// 		env, err := utils.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
// 		if err != nil {
// 			fmt.Printf("GetEnvelopeFromBlock returned error: %s", err)
// 		}

// 		channelHeader, err := utils.ChannelHeader(env)
// 		if err != nil {
// 			fmt.Printf("ChannelHeader returned error: %s", err)
// 		}
// 		typeCode := channelHeader.GetType()
// 		typeInfo := fabricutils.TypeCodeToInfo(typeCode)
// 		createdAt := time.Unix(channelHeader.GetTimestamp().Seconds, int64(channelHeader.GetTimestamp().Nanos))

// 		// Checking validity
// 		txsFltr := util.TxValidationFlags(blockResponse.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])

// 		// Processing transactions
// 		for i, d := range blockResponse.Data.Data {
// 			fmt.Printf("tx %d (validation status: %s):\n", i, txsFltr.Flag(i).String())

// 			env, err := utils.GetEnvelopeFromBlock(d)
// 			if err != nil {
// 				return 0, err
// 			}

// 			payload, err := utils.GetPayload(env)
// 			if err != nil {
// 				return 0, err
// 			}
// 			chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
// 			if err != nil {
// 				return 0, err
// 			}

// 			shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
// 			if err != nil {
// 				return 0, err
// 			}

// 			tx, err := utils.GetTransaction(payload.Data)
// 			if err != nil {
// 				return 0, err
// 			}

// 			if typeInfo != "ENDORSER_TRANSACTION" {
// 				event := beat.Event{
// 					Timestamp: time.Now(),
// 					Fields: libbeatCommon.MapStr{
// 						"type":             b.Info.Name,
// 						"block_number":     blockResponse.Header.Number,
// 						"tx_id":            chdr.TxId,
// 						"channel_id":       chdr.ChannelId,
// 						"index_name":       bt.config.TransactionIndexName,
// 						"created_at":       createdAt,
// 						"creator":          fabricutils.ReturnCreatorString(shdr.Creator),
// 						"transaction_type": typeInfo,
// 					},
// 				}
// 				beatClient.Publish(event)
// 				logp.Info("Config transaction event sent")
// 			} else {
// 				_, respPayload, payloadErr := utils.GetPayloads(tx.Actions[0])
// 				if payloadErr != nil {
// 					return 0, payloadErr
// 				}

// 				txRWSet := &rwsetutil.TxRwSet{}
// 				err = txRWSet.FromProtoBytes(respPayload.Results)
// 				if err != nil {
// 					return 0, err
// 				}

// 				readset := []*fabricutils.Readset{}
// 				writeset := []*fabricutils.Writeset{}
// 				// Getting read-write set
// 				readIndex := 0
// 				writeIndex := 0
// 				// For every namespace
// 				for _, ns := range txRWSet.NsRwSets {

// 					if len(ns.KvRwSet.Writes) > 0 {
// 						// Getting the writes
// 						for _, w := range ns.KvRwSet.Writes {
// 							writeset = append(writeset, &fabricutils.Writeset{})
// 							writeset[writeIndex].Namespace = ns.NameSpace
// 							writeset[writeIndex].Key = w.Key
// 							writeset[writeIndex].Value = string(w.Value)
// 							writeset[writeIndex].IsDelete = w.IsDelete
// 							writeIndex++

// 							// Sending a new event to the "key" index with the write data
// 							event := beat.Event{
// 								Timestamp: time.Now(),
// 								Fields: libbeatCommon.MapStr{
// 									"type":              b.Info.Name,
// 									"tx_id":             chdr.TxId,
// 									"channel_id":        chdr.ChannelId,
// 									"chaincode_name":    respPayload.ChaincodeId.Name,
// 									"chaincode_version": respPayload.ChaincodeId.Version,
// 									"index_name":        bt.config.KeyIndexName,
// 									"key":               w.Key,
// 									"value":             string(w.Value),
// 									"created_at":        createdAt,
// 									"creator":           fabricutils.ReturnCreatorString(shdr.Creator),
// 								},
// 							}
// 							beatClient.Publish(event)
// 							logp.Info("Write event sent")
// 						}
// 					}

// 					if len(ns.KvRwSet.Reads) > 0 {
// 						// Getting the reads
// 						for _, w := range ns.KvRwSet.Reads {
// 							readset = append(readset, &fabricutils.Readset{})
// 							readset[readIndex].Namespace = ns.NameSpace
// 							readset[readIndex].Key = w.Key
// 							readIndex++
// 						}
// 					}
// 				}
// 				transactions = append(transactions, chdr.TxId)
// 				// Sending the transaction data to the "transaction" index
// 				event := beat.Event{
// 					Timestamp: time.Now(),
// 					Fields: libbeatCommon.MapStr{
// 						"type":              b.Info.Name,
// 						"block_number":      blockResponse.Header.Number,
// 						"tx_id":             chdr.TxId,
// 						"channel_id":        chdr.ChannelId,
// 						"chaincode_name":    respPayload.ChaincodeId.Name,
// 						"chaincode_version": respPayload.ChaincodeId.Version,
// 						"index_name":        bt.config.TransactionIndexName,
// 						"created_at":        createdAt,
// 						"creator":           fabricutils.ReturnCreatorString(shdr.Creator),
// 						"readset":           readset,
// 						"writeset":          writeset,
// 						"transaction_type":  typeInfo,
// 					},
// 				}
// 				beatClient.Publish(event)
// 				logp.Info("Endorsement transaction event sent")
// 			}

// 		}

// 		prevHash := hex.EncodeToString(blockResponse.Header.PreviousHash)
// 		dataHash := hex.EncodeToString(blockResponse.Header.DataHash)
// 		blockHash := fabricutils.GenerateBlockHash(blockResponse.Header.PreviousHash, blockResponse.Header.DataHash, blockResponse.Header.Number)
// 		// Sending the block data to the "block" index
// 		event := beat.Event{
// 			Timestamp: time.Now(),
// 			Fields: libbeatCommon.MapStr{
// 				"type":          b.Info.Name,
// 				"block_number":  blockResponse.Header.Number,
// 				"block_hash":    blockHash,
// 				"previous_hash": prevHash,
// 				"data_hash":     dataHash,
// 				"created_at":    createdAt,
// 				"index_name":    bt.config.BlockIndexName,
// 				"transactions":  transactions,
// 			},
// 		}
// 		beatClient.Publish(event)
// 		logp.Info("Block event sent")

// 		// Send the latest known block number to Elasticsearch
// 		jsonBlockNumber, err := json.Marshal(lastBlockNumber)
// 		if err != nil {
// 			return 0, err
// 		}
// 		resp, err := http.Post(fmt.Sprintf(bt.Fsetup.ElasticURL+"/last_block_%s_%d/_doc/1", bt.config.Peer, index), "application/json", bytes.NewBuffer(jsonBlockNumber))
// 		if err != nil {
// 			return 0, err
// 		}
// 		defer resp.Body.Close()
// 		if resp.StatusCode != 200 && resp.StatusCode != 201 {
// 			return 0, errors.New("Sending last block number to Elasticsearch failed!")
// 		}
// 		blockNumber++
// 		lastBlockNumber.BlockNumber++
// 	}

// 	return blockNumber, nil
// }
