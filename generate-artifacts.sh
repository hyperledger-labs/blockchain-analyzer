#!/bin/bash

export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
export FABRIC_CFG_PATH=$PWD

CRYPTOGEN=./cryptogen
CONFIGTXGEN=./configtxgen

WHICHOS=`uname -a`;
OSTYPE="Linux"
echo $WHICHOS
if [[  "${WHICHOS}" == *"Darwin"* ]]; then
  OSTYPE="Darwin"
fi

if [[ ${OSTYPE} == "Darwin" ]]; then
  CRYPTOGEN=./cryptogen_mac
  CONFIGTXGEN=./configtxgen_mac
fi

#Generate crypto material using crypto-config.yaml as config file
${CRYPTOGEN} generate --config=./crypto-config.yaml

#Rename admin private key files so that their names are always the same (no need to change Hyperledger Explorer configuration after restarting the network)
for ORG_NUM in 1 2 3 4
do
	mv ./crypto-config/peerOrganizations/org$ORG_NUM.el-network.com/users/Admin@org$ORG_NUM.el-network.com/msp/keystore/*_sk ./crypto-config/peerOrganizations/org$ORG_NUM.el-network.com/users/Admin@org$ORG_NUM.el-network.com/msp/keystore/adminKey$ORG_NUM
done
mv ./crypto-config/ordererOrganizations/el-network.com/users/Admin@el-network.com/msp/keystore/*_sk ./crypto-config/ordererOrganizations/el-network.com/users/Admin@el-network.com/msp/keystore/ordererAdminKey

#Generate configuration txs
mkdir channel-artifacts
${CONFIGTXGEN} -profile FourOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
export CHANNEL_NAME=mychannel
${CONFIGTXGEN} -profile FourOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME

${CONFIGTXGEN} -profile FourOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
${CONFIGTXGEN} -profile FourOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP
${CONFIGTXGEN} -profile FourOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org3MSP
${CONFIGTXGEN} -profile FourOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org4MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org4MSP
