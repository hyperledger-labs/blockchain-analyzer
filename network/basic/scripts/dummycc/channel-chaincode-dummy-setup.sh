#!/bin/bash

export CHANNEL_NAME=mychannel

#Create channel
export CACERT_ORDERER=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/el-network.com/orderers/orderer.el-network.com/tls/ca.crt
export CACERT_1=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.el-network.com/peers/peer0.org1.el-network.com/tls/ca.crt
export CACERT_2=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.el-network.com/peers/peer0.org2.el-network.com/tls/ca.crt
export CACERT_3=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.el-network.com/peers/peer0.org3.el-network.com/tls/ca.crt
export CACERT_4=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org4.el-network.com/peers/peer0.org4.el-network.com/tls/ca.crt
peer channel create -o orderer.el-network.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls --cafile $CACERT_ORDERER

#Connect peers to channel
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
    eval "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org$ORG_NUM.el-network.com/users/Admin@org$ORG_NUM.el-network.com/msp CORE_PEER_ADDRESS=peer0.org$ORG_NUM.el-network.com:7051 CORE_PEER_LOCALMSPID=Org$(($ORG_NUM))MSP CORE_PEER_TLS_ROOTCERT_FILE=$CACERT"

    echo "Connecting peer$ORG_NUM to channel..."
    peer channel join -b mychannel.block
done

#Update channel
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
    eval "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org$ORG_NUM.el-network.com/users/Admin@org$ORG_NUM.el-network.com/msp CORE_PEER_ADDRESS=peer0.org$ORG_NUM.el-network.com:7051 CORE_PEER_LOCALMSPID=Org$(($ORG_NUM))MSP CORE_PEER_TLS_ROOTCERT_FILE=$CACERT"
    echo "Updating channel..."
    peer channel update -o orderer.el-network.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/Org$(($ORG_NUM))MSPanchors.tx --tls --cafile $CACERT_ORDERER
done

#Install chaincode
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
    eval "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org$ORG_NUM.el-network.com/users/Admin@org$ORG_NUM.el-network.com/msp CORE_PEER_ADDRESS=peer0.org$ORG_NUM.el-network.com:7051 CORE_PEER_LOCALMSPID=Org$(($ORG_NUM))MSP CORE_PEER_TLS_ROOTCERT_FILE=$CACERT"
    echo "Installing chaincode on peer$ORG_NUM..."
    peer chaincode install -n dummycc -v 5.5 -l node -p /opt/gopath/src/github.com/chaincode/dummycc
done

# Instantiate chaincode
CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.el-network.com/users/Admin@org1.el-network.com/msp
CORE_PEER_ADDRESS=peer0.org1.el-network.com:7051
CORE_PEER_LOCALMSPID=Org1MSP
CORE_PEER_TLS_ROOTCERT_FILE=$CACERT_1

peer chaincode instantiate -o orderer.el-network.com:7050 -C $CHANNEL_NAME -n dummycc -l node -v 5.5 -c '{"Args":[]}' -P "OR ('Org1MSP.member','Org2MSP.member','Org3MSP.member','Org4MSP.member')" --tls --cafile $CACERT_ORDERER
