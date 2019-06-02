#!/bin/bash

export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
export FABRIC_CFG_PATH=$PWD
export CHANNEL_NAME=mychannel

COMPOSE_PROJECT_NAME=$CHANNEL_NAME docker-compose -f docker-compose-cli.yaml up -d

docker exec -it cli bash
