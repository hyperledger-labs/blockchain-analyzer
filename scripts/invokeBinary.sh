#!/bin/bash

set -e

export CHANNEL_NAME=mychannel
export CACERT_ORDERER=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/tls/ca.crt

if [ $# == 2 ];
    then
        echo "Creating $1 with value $2"
        peer chaincode invoke -o orderer.el-network.com:7050 -C $CHANNEL_NAME -n mycc -c "{\"Args\":[\"set\",\"$1\",\"$2\"]}" --tls --cafile $CACERT_ORDERER
        echo "Created"
    else
        echo "Wrong number of arguments! Usage: invokeBinary.sh <key, value(int)>"
fi
