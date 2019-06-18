#!/bin/bash

export CHANNEL_NAME=mychannel
export CACERT_ORDERER=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/tls/ca.crt

peer chaincode invoke -o orderer.el-network.com:7050 -C $CHANNEL_NAME -n go -c "{\"Args\":[\"initLedger\"]}" --tls --cafile $CACERT_ORDERER

