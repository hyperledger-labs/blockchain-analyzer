package fabricutils

import (
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"fmt"
	"math"

	"github.com/blockchain-analyzer/agent/agentmodules/fabricsetup"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type asn1Header struct {
	Number       int64
	PreviousHash []byte
	DataHash     []byte
}

// Generates block hash from previous hash, data hash and block number.
func GenerateBlockHash(previousHash, dataHash []byte, blockNumber uint64) string {

	h := common.BlockHeader{
		Number:       blockNumber,
		PreviousHash: previousHash,
		DataHash:     dataHash}

	asn1Header := asn1Header{
		PreviousHash: h.PreviousHash,
		DataHash:     h.DataHash,
	}
	if h.Number > uint64(math.MaxInt64) {
		panic(fmt.Errorf("Golang does not currently support encoding uint64 to asn1"))
	} else {
		asn1Header.Number = int64(h.Number)
	}
	asn1Bytes, err := asn1.Marshal(asn1Header)
	if err != nil {
		// Errors should only arise for types which cannot be encoded, since the
		// BlockHeader type is known a-priori to contain only encodable types, an
		// error here is fatal and should not be propogated
		panic(err)
	}

	return hex.EncodeToString(util.ComputeSHA256(asn1Bytes))
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

// This function is borrowed from an opensource project: https://github.com/ale-aso/fabric-examples/blob/master/blockparser.go
// It returns the creator certificate for the specified transaction.
func ReturnCreatorString(bytes []byte) string {
	defaultString := strings.Replace(string(bytes), "\n", ".", -1)

	sId := &msp.SerializedIdentity{}
	err := proto.Unmarshal(bytes, sId)
	if err != nil {
		return defaultString
	}

	bl, _ := pem.Decode(sId.IdBytes)
	if bl == nil {
		return defaultString
	}

	return string(sId.IdBytes)
}

func ReturnCreatorOrgString(bytes []byte) string {
	defaultString := strings.Replace(string(bytes), "\n", ".", -1)

	sId := &msp.SerializedIdentity{}
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

func IndexOfChaincode(array []fabricsetup.Chaincode, name string) int {
	for i, v := range array {
		if v.Name == name {
			return i
		}
	}
	return -1
}

type Readset struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type Writeset struct {
	Namespace string      `json:"namespace"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	IsDelete  bool        `json:"isDelete"`
}
