#!/bin/bash

set -e

export CHANNEL_NAME=mychannel
export CACERT_ORDERER=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/tls/ca.crt

if [ $# == 1 ];
    then
        peer chaincode query -o orderer.el-network.com:7050 -C $CHANNEL_NAME -n mycc -c "{\"Args\":[\"get\",\"$1\"]}" --tls --cafile $CACERT_ORDERER
    else
        echo "Wrong number of arguments! Usage: queryBinary.sh <key>"
fi
