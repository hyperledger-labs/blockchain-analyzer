#!/bin/bash

export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
export FABRIC_CFG_PATH=$PWD
export CHANNEL_NAME=applechannel

echo "Stopping all containers"

#docker rm -f $(docker ps -aq)
COMPOSE_PROJECT_NAME=$CHANNEL_NAME docker-compose down
docker rm -f $(docker ps -aq) 2>/dev/null

echo "Removing channel artifacts and generated crypto material..."

source ./destroy-artifacts.sh