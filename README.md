# Hyperledger Summer Internship: Analyzing Hyperledger Fabric Ledger, Transactions and Logs using Elasticsearch and Kibana

Mentor: Salman Baset [salmanbaset](https://github.com/salmanbaset)

Mentee: Balazs Prehoda [balazsprehoda](https://github.com/balazsprehoda)

## Contents
1. [Description](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#description)  
2. [Expected Outcome](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#expected-outcome)  
3. [Overview](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#overview)
4. [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#prerequisites)  
5. [Getting Started](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#getting-started)  
6. [Fabric Network](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#fabric-network)  
  6.1. [Basic network](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#basic-network)  
  6.2. [Multichannel network](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#multichannel-network)  
  6.3. [(Optional) Generate makefile configuration](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#optional-generate-makefile-configuration)  
  6.4. [Start the network](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#start-the-network)   
  6.5. [Stop the network](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#stop-the-network)  
7. [Dummyapp Application](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#dummyapp-application)  
  7.1. [Installation](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#installation)  
  7.2. [Configuration](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#configuration)  
  7.3. [User enrollment and registration](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#user-enrollment-and-registration)  
  7.4. [Invoke transactions](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#invoke-transactions)  
  7.5. [Query](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#query)  
8. [Elastic Stack](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#elastic-stack)  
  8.1. [Credit](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#credit)  
  8.2. [Description](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#description-1)  
  8.3. [Start](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#start)  
  8.4. [Stop and destroy](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#stop-and-destroy)  
9. [Beats Agent](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#beats-agent)  
  9.1. [Environment setup](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#environment-setup)  
  9.2. [Configure fabricbeat](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#configure-fabricbeat)  
  9.3. [Build fabricbeat](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#build-fabricbeat)  
  9.4. [Start fabricbeat](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#start-fabricbeat)   
  9.5. [Stop fabricbeat](https://github.com/balazsprehoda/hyperledger-elastic/tree/test-build-path#stop-fabricbeat)  
  

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

## Overview

Basic data flow: 
![alt text](https://github.com/balazsprehoda/hyperledger-elastic/blob/test-build-path/Fabric%20Network%20And%20Fabricbeat.jpg "Fabric Network And Fabricbeat Basic Data Flow")

## Prerequisites

Please make sure that you have set up the environment for the project. Follow the steps listed in [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/blob/test-build-path/Prerequisites.md).   

## Getting Started

To get started with the project, clone the git repository. It is important that you place it under $GOPATH/src/github.com  
```
cd $GOPATH/src/github.com  
git clone https://github.com/balazsprehoda/hyperledger-elastic.git  
```

## Fabric Network

The `network` directory contains two network setups, `basic` and `multichannel`. 

### Basic network

It is a simple test network with 4 organizations, 1 peer each, a solo orderer communicating over TLS and a sample chaincode called `dummycc`. It writes deterministically generated hashes and (optionally) previous keys as value to the ledger.  

### Multichannel network

It is a test network setup with 4 organizations, 2 peers each, a solo orderer communicating over TLS and two channels:

1. `fourchannel`:
  * members: all four organizations
  * chaincode: `dummycc`: It writes deterministically generated hashes and (optionally) previous keys as value to the ledger.
2. `twochannel`:
  * members: only `Org1` and `Org3`
  * chaincode: `fabcar`: The classic fabcar example chaincode extended with a `getHistoryForCar()` chaincode function.

### (Optional) Generate makefile configuration
```
make generate ABSPATH={ABSOLUTE_PATH_TO_SRC_DIR}
```

### Start the network

To generate crypto and setup the network, make sure that the `GOPATH` variable is set correctly. After that, set `ABSPATH` to point to the parent folder of `hyperledger-elastic` (e.g., `export ABSPATH=$GOPATH/src/github.com`). When we are done with setting these environment variables, we can start the network. Issue the following command in the `network/basic` directory to start the basic network, or in the `network/multichannel` directory to start the multichannel network:

```
make start  
```

We can enter the CLI by issuing the command 
```
docker exec -it cli bash  
```

Inside the CLI, the `/scripts` folder contains the scripts that can be used to install, instantiate and invoke chaincode (though the `make start` command takes care of installation and instantiation).

### Stop the network

To stop the network and delete all the generated data (crypto material, channel artifacts and dummyapp wallet), run

```
make destroy
```

## Dummyapp Application  
The dummyapp application is used to generate users and transactions. It can connect to both the basic and the multichannel networks.  
The commands in this section should be issued from the `hyperledger-elastic/apps/dummyapp` directory.

### Installation
Before the first run, we have to install the necessary node modules:
```
npm install
```

### Configuration
The `config.json` contains the configuration for the application. We can configure the channel and chaincode name that we want our application to use, the users we want to enroll and the transactions we want to initialize. Transactions have 4 fields:
1. `user`: This field is required. We have to specify which user to use when making the transaction.
2. `txFunction`: This field is required. We have to specify here the chaincode function that should be called.
3. `key`: This field is required. We have to specify here the key to be written to the ledger.
4. `previousKey`: This field is optional. We can specify here the key to which the new transaction (key-value pair) is linked.

###  User enrollment and registration

Setup path of the network configuration (basic or multichannel shipped with this repository)
```
export ABSPATH={ABSOLUTE_PATH}/src/github.com/hyperledger-elastic/network/basic/
```

To enroll admins, register and enroll users, run the following command:
```
make users
```

###  Invoke transactions
To add key-value pairs, run
```
make invoke
```

### Query
To query a specific key, run
```
make query KEY=key1
```
To query all key-value pairs, run
```
make query-all
```

## Elastic Stack
This project includes an Elasticsearch and Kibana setup to index and visualize blockchain data.  
The commands in this section should be issued from the `hyperledger-elastic/stack` folder.

### Credit
This setup is borrowed from
https://github.com/maxyermayank/docker-compose-elasticsearch-kibana

### Description
Kibana container and Elasticsearch cluster with nginx and 3 Elasticsearch containers. To view Kibana in browser, navigate to http://localhost:5601

### Start
To start the containers, issue
```
make start
```

### Stop and destroy
To stop the containers, issue
```
make destroy
```
To stop the containers and remove old data, run
```
make erase
```

## Beats Agent

The fabricbeat beats agent is responsible for connecting to a specified peer, periodically querying its ledger, processing the data and shipping it to Elasticsearch. Multiple instances can be run at the same time, each querying a different peer and sending its data to the Elasticsearch cluster.  
The commands in this section should be issued from the `hyperledger-elastic/agent/fabricbeat` directory.

### Environment setup

Before configuring and building the fabricbeat agent, we should make sure that the `GOPATH` variable is set correctly. Then, we have to add `$GOPATH/bin` to the `PATH`:
```
export PATH=$PATH:$GOPATH/bin
```
After that, we have to set the `BEAT_PATH` variable to point to the fabricbeat folder:
```
export BEAT_PATH=$GOPATH/src/github.com/hyperledger-elastic/agent/fabricbeat
```
Note: there is no trailing slash.
We use vendoring instead of go modules, so we have to make sure `GO111MODULE` is set to `auto` (it is the default):  
```
export GO111MODULE=auto
```  

### Configure fabricbeat

We can configure the agent using the `fabricbeat.yml` file. This file is generated based on the `fabricbeat/_meta/beat.yml` file. If we want to update the generated config file, we can edit `fabricbeat/_meta/beat.yml`, then run
```
make update
```

The configurable fields are the following:
* `period`: defines how often an event is sent to the output (Elasticsearch in this case)
* `organization`: defines which organization the connected peer is part of
* `peer`: defines the peer which fabricbeat should query (must be defined in the connection profile)
* `connectionProfile`: defines the location of the connection profile of the Fabric network
* `adminCertPath`: absolute path to the admin certfile
* `adminKeyPath`: absolute path to the admin keyfile
* `elasticURL`: URL of Elasticsearch (defaults to http://localhost:9200)
* `kibanaURL`: URL of Kibana (defaults to http://localhost:5601)
* `blockIndexName`: defines the name of the index to which the block data should be sent
* `transactionIndexName`: defines the name of the index to which the transaction data should be sent
* `keyIndexName`: defines the name of the index to which the key write data should be sent
* `dashboardDirectory`: folder which should contain the generated dashboards
* `templateDirectory`: folder which contains the templates for Kibana objects (index patterns, dashboards, etc.)
* `chaincodes`: describes the chaincodes installed on the peer
  * `name`: the name of the chaincode
  * `values`: the keys of the values that get persisted with the key (e.g. fabcar: key: CAR0 values: [make, model, colour, owner])
  * `linkingKey`: the name of the key that links transactions (e.g. dummycc: previousKey)
* `setup.ilm.enabled`: setting this false makes possible to define our own indices (for blocks, transactions and keys per organization)
* `output.elasticsearch.index`: the template for runtime index creation
* `output.elasticsearch.hosts`: the list of elasticsearch hosts we want our agent to connect to
* `setup.template.name`: the name of the index template that is going to be automatically created if does not exist
* `setup.template.pattern`: the index template is loaded for indices matching this pattern
* `setup.dashboards.directory`: the directory that contains the (generated) dashboards to be imported into Kibana on start of the agent

The paths and peer/org names contain variables that can be passed when starting the agent:
* `GOPATH` is the value of the gopath environment variable
* `ORG_NUMBER` is the number of the organization (1 for Org1, ..., 4 for Org4)
* `NETWORK` is the name of the network (basic or multichannel)
If we want to use the agent with another (custom) network, we have to modify the configuration according to the network's specifications. We can remove and ignore these variables and hardcode the names and paths. We can run multiple instances at the same time with different configurations (the workflow for this scenario is 1. modify config 2. `make update` 3. `make` 4. run the agent 5. modify config 6. `make update` 7. run another instance of the agent).

### About indices

We use 3 different indices per organization: one for blocks, one for transactions and one for single writes.  
If we run multiple agents for peers in the same organization, they are goint to send their data to the same indices. We can then select the peer on the dashboards to view its data only.  
If we run multiple instances for peers in different organizations, we are going to see the data of different organizations on different dashboards.  
The name of the indices can be customized in the fabricbeat configuration file (\_meta/beat.yml and `make update` or directly in fabricbeat.yml).

### Build fabricbeat

To build the agent, issue the following command:
```
make
```

### Start fabricbeat

To start the agent, issue the following command from the `fabricbeat` directory:
```
./fabricbeat -e -d "*"
```
If we want to use the agent with one of the example networks from the `hyperledger-elastic/network` folder, we can start the agent using:
```
ORG_NUMBER=1 NETWORK=basic ./fabricbeat -e -d "*"
```
The variables passed are used in the configuration (`fabricbeat.yml`). To connect to another network or peer, change the configuration (and/or the passed variables) accordingly.

### Stop fabricbeat

To stop the agent, simply type `Ctrl+C`
