# Hyperledger Summer Internship: Analyzing Hyperledger Fabric Ledger, Transactions and Logs using Elasticsearch and Kibana

Mentor: Salman Baset [salmanbaset](https://github.com/salmanbaset)

Mentee: Balazs Prehoda [balazsprehoda](https://github.com/balazsprehoda)

## Description

Each blockchain platform, including Hyperledger Fabric, provide a way to record information on blockchain in an immutable manner. In the case of Hyperledger Fabric, information is recorded as a `key-value` pair. All previous updates to a `key` are recorded in the ledger, but only the latest value of a `key` can be easily queried using CouchDB; the previous updates are only available in ledger files. This mechanism makes it challenging to perform analysis of updates to a `key`, a necessary requirement for information provenance.

The goal of this project is to

1. write a Elastic beats module (in Go), that will ship ledger data to Elasticsearch instance

2. create generic Kibana dashboards that will allow selection of a particular key and visualization updates to it (channel, id, timestamp etc)

Time permitting, the dashboards can be extended to analyze Fabric logs and in-progress transaction data, as well as creating dashboards similar to Hyperledger Explorer.

Of course, a blockchain solution can track information provenance in multiple ways. In one such mechanism, a solution may always write new key-value pairs to blockchain, and maintain the relationship among key-value pairs within the solution (off-chain), instead of blockchain. This project does not concern itself on how a solution manages relationship among key-value pairs.

## Expected Outcome

A open source implementation, eventually available as Hyperledger Labs, containing:

Elastic beats plugin for Hyperledger Fabric
Kibana dashboards
Dashboards similar to Hyperledger Explorer
Create a setup for generating various dummy data in various configurations
One peer / CA / order, single user for initial testing
A four peers/CA setup with two channels, and two users each associated with two peers. Select (e.g.) 10 keys (through configuration file), to which these users write data, for at least one value per key.

## Fabric network

It is a simple test network with 4 organizations, 1 peer each, a solo orderer communicating over TLS and the basic Fabcar chaincode extended with a `getHistoryForCar()` function. Hyperledger Explorer can be connected using the `connectionProfile.json` file.

To generate crypto and setup the network, simply run `start-network.sh`
To stop the network and delete all the generated data (crypto material and channel artifacts), run `destroy-network.sh`

Inside the CLI, the `/scripts` folder contains the scripts that can be used to install, instantiate and invoke chaincode.
