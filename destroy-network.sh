#!/bin/bash

echo "Removing channel artifacts and generated crypto material..."

source ./destroy-artifacts.sh

echo "Channel artifacts and crypto material removed"

echo "Stopping all containers"

sudo docker rm -f $(docker ps -aq)

echo "All containers stopped"
