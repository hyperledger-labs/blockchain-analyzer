package ledgerutils

import (
	"time"

	protoCommon "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/blockchain-analyzer/agent/agentmodules/fabricutils"

	"log"
)

func GetBlockHash(blockNumber uint64, ledgerClient *ledger.Client) (string, error) {
	blockResponse, blockError := ledgerClient.QueryBlock(blockNumber)
	if blockError != nil {
		return "", blockError
	}
	log.Print("Querying last known block from ledger successful")

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
	env, err := protoutil.GetEnvelopeFromBlock(blockResponse.Data.Data[0])
	if err != nil {
		return nil, "", time.Now(), nil, err
	}

	channelHeader, err := protoutil.ChannelHeader(env)
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

func ProcessTx(txData []byte) (txId, channelId, creator, creatorOrg string, tx *peer.Transaction, err error) {
	env, err := protoutil.GetEnvelopeFromBlock(txData)
	if err != nil {
		return "", "", "", "", nil, err
	}

	payload, err := protoutil.UnmarshalPayload(env.GetPayload())
	if err != nil {
		return "", "", "", "", nil, err
	}

	chdr, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return "", "", "", "", nil, err
	}

	shdr, err := protoutil.UnmarshalSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		return "", "", "", "", nil, err
	}

	tx, err = protoutil.UnmarshalTransaction(payload.Data)
	if err != nil {
		return "", "", "", "", nil, err
	}

	return chdr.TxId, chdr.ChannelId, fabricutils.ReturnCreatorString(shdr.Creator), fabricutils.ReturnCreatorOrgString(shdr.Creator), tx, nil
}

func ProcessEndorserTx(txData []byte) (txId, channelId, creator, creatorOrg string, txRWSet *rwsetutil.TxRwSet, chaincodeName, chaincodeVersion string, err error) {

	txId, channelId, creator, creatorOrg, tx, err := ProcessTx(txData)
	if err != nil {
		return "", "", "", "", nil, "", "", err
	}

	_, respPayload, payloadErr := protoutil.GetPayloads(tx.Actions[0])
	if payloadErr != nil {
		return "", "", "", "", nil, "", "", err
	}

	txRWSet = &rwsetutil.TxRwSet{}
	err = txRWSet.FromProtoBytes(respPayload.Results)
	if err != nil {
		return "", "", "", "", nil, "", "", err
	}

	return txId, channelId, creator, creatorOrg, txRWSet, respPayload.ChaincodeId.Name, respPayload.ChaincodeId.Version, nil
}
