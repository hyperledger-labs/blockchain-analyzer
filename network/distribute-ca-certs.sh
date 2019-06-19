#!/bin/bash

export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN

# Get CA certificate for every peer and orderer
fabric-ca-client getcainfo -u http://localhost:7054 -M $PWD/crypto-config/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/msp

for ORG_NUM in 1 2 3 4
do
if [ $ORG_NUM == 1 ]; then
        CACERT=$CACERT_1
    else if [ $ORG_NUM == 2 ]; then
            CACERT=$CACERT_2
        else if [ $ORG_NUM == 3 ]; then
                CACERT=$CACERT_3
            else
                CACERT=$CACERT_4
        fi
    fi
fi
    fabric-ca-client getcainfo -u http://localhost:7054 -M $PWD/crypto-config/peerOrganizations/org${ORG_NUM}.el-network.com/peers/peer0.org${ORG_NUM}.el-network.com/msp
done