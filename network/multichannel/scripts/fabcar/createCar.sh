#!/bin/bash

export CHANNEL_NAME=mychannel
export CACERT_ORDERER=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/tls/ca.crt

if [ $# == 5 ];
    then
        peer chaincode invoke -o orderer.el-network.com:7050 -C $CHANNEL_NAME -n go -c "{\"Args\":[\"createCar\",\"$1\",\"$2\",\"$3\",\"$4\",\"$5\"]}" --tls --cafile $CACERT_ORDERER
    else
        echo "Usage: createCar.sh <id, make, model, colour, owner>"
fi
