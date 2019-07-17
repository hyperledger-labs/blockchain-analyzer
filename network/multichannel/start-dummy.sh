#!/bin/bash

export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
export FABRIC_CFG_PATH=$PWD
# export CHANNEL_NAME=mychannel

source ./generate-artifacts.sh

# COMPOSE_PROJECT_NAME=$CHANNEL_NAME docker-compose -f docker-compose.yaml up -d
COMPOSE_PROJECT_NAME=multichannel docker-compose -f docker-compose.yaml up -d

sleep 5
docker exec -it cli scripts/dummycc/channel-chaincode-dummy-setup.sh
