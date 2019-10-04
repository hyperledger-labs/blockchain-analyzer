# Getting Started with Multichannel Network

This is an example to setup the project with `multichannel` network on a new Ubuntu 18.04 / 16.04 virtual machine from scratch. The instructions should also work on Mac OS.

1. [Install Prequisites](#install-prerequisites)
2. [Clone the repository](#clone-the-repository)
3. [Start / stop a Hyperledger Fabric network using `multichannel` configuration](#start--stop-the-multichannel-network). See `multichannel` page for multi-channel configuration.
4. [Create users and transactions using dummy application](#create-users-and-transactions-using-dummy-application)
5. [Start Elastic stack](#start-elastic-stack)
6. [Build fabricbeat agent](#build-fabricbeat-agent)
7. [Start the fabricbeat agent](#start-fabricbeat) and connect to peer in `basic` network
8. [Configuring Indices for the first time in Kibana](#configuring-indices-for-the-first-time-in-Kibana)
9. [Viewing dashboards](#viewing-dashboards) that store data
10. [Starting more instances of fabricbeat agent](#starting-more-instances-of-fabricbeat-agent)

## Install Prerequisites

Please make sure that you have set up the environment for the project. Follow the steps listed in [Prerequisites](https://github.com/balazsprehoda/hyperledger-elastic/blob/master/docs/Prerequisites.md).   

## Clone the repository
To get started with the project, clone the git repository. It is important that you place it under `$GOPATH/src/github.com`  
```
$ mkdir $GOPATH/src/github.com -p
$ cd $GOPATH/src/github.com  
$ git clone https://github.com/hyperledger-labs/blockchain-analyzer.git
```

## Start / stop the `multichannel` network
It is a test network setup with four organizations, two peers per organization, a solo orderer communicating over TLS and two channels:

1. `fourchannel`:
  * members: all four organizations
  * chaincode: `dummycc`: It writes deterministically generated hashes and (optionally) previous keys as value to the ledger.
2. `twochannel`:
  * members: only `Org1` and `Org3`
  * chaincode: `fabcar`: The classic fabcar example chaincode extended with a `getHistoryForCar()` chaincode function.


The sample chaincode called `dummycc` (used in basic) also works with multichannel. It writes 
deterministically generated hashes and (optionally) previous keys as value to the ledger.  


Issue the following command in the `network/multichannel` directory
```
make start
```

Enter the Fabric CLI Docker container by issuing the command:
```
docker exec -it cli bash  
```

Inside the CLI, the `/scripts` folder contains the scripts that can be used to install, instantiate and invoke chaincode (though the `make start` command takes care of installation and instantiation).

### Stop the network

To stop the network and delete all the generated data (crypto material, channel artifacts 
and dummyapp wallet), run:

```
make destroy
```

## Create users and transactions using dummy application

The dummyapp application is used to create users and generate transactions
for different scenarios so that the resulting transactions can be 
analyzed with the Elastic stack. Examples of scenarios include which 
channels to use, and which Fabric ca users to invoke transactions.

This application can connect to both the basic and the multichannel networks.  

The commands in this section should be issued from the `blockchain-analyzer/apps/dummyapp` directory.

### Install dummy application dependencies
Before the first run, we have to install the necessary node modules:
```
npm install
```

Take a look at the `config.json` file. This file contains the transactions that will be invoked into the `basic` network.


###  User enrollment and registration
To enroll admins, register and enroll users, run the following command:
```
NETWORK=multichannel CHANNEL=fourchannel make users
```

###  Invoke transactions
To add key-value pairs, run:
```
NETWORK=multichannel CHANNEL=fourchannel make invoke
```

### Query
To query a specific key, run
```
NETWORK=multichannel CHANNEL=fourchannel make query KEY=Key1
```
To query all key-value pairs, run
```
NETWORK=multichannel CHANNEL=fourchannel make query-all
```


## Start Elastic stack

This project includes an Elasticsearch and Kibana setup to index and visualize blockchain data.  
The commands in this section should be issued from the `blockchain-analyzer/stack` folder.

If you are working in a machine with low memory, the Elasticsearch container may not start. In this case, issue the following command:
```
sudo sysctl -w vm.max_map_count=262144
```
to set the vm.max_map_count kernel setting to 262144, then destroy and bring up the Elastic Stack again.

### Credit
This setup is borrowed from
https://github.com/maxyermayank/docker-compose-elasticsearch-kibana

### Start
To start the containers, navigate to `blockchain-analyzer/stack` directory and issue:
```
make start
```

### View Kibana
To view Kibana in browser, navigate to http://localhost:5601 . It can take some time (2-5 minutes) for Kibana to start depending on your machine configuration.

### Stop and destroy
To stop the containers, issue
```
make destroy
```
To stop the containers and remove old data, run
```
make erase
```


## Build Fabricbeat Agent

The fabricbeat beats agent is responsible for connecting to a specified peer, periodically querying its ledger, processing the data and shipping it to Elasticsearch. Multiple instances can be run at the same time, each querying a different peer and sending its data to the Elasticsearch cluster.  
The commands in this section should be issued from the `blockchain-analyzer/agent/fabricbeat` directory.

### Environment setup

Before configuring and building the fabricbeat agent, please make sure that the `GOPATH` variable is set correctly. Then, add `$GOPATH/bin` to the `PATH`:
```
export PATH=$PATH:$GOPATH/bin
```
Ensure that Python version is 2.7.*. 

Get module dependencies:
```
make get-go
```

Build the agent:
```
make update
make
```

## Start fabricbeat

To start the agent, issue the following command from the `fabricbeat` directory:
```
./fabricbeat -e -d "*"
```
To use the agent with the `multichannel` network from the `blockchain-analyzer/network` folder, you can start the agent using:
```
ORG_NUMBER=1 PEER_NUMBER=0 NETWORK=multichannel ./fabricbeat -e -d "*"
```
The variables passed are used in the configuration (`fabricbeat.yml`). To connect to another network or peer, change the configuration (and/or the passed variables) accordingly.

### Stop fabricbeat

To stop the agent, simply type `Ctrl+C`


## Configuring Indices for the first time in Kibana

Next, we navigate to http://localhost:5601.

Click the dashboards icon on the left:
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Starting_page.png "Kibana starting page")


Kibana takes us to select a default index pattern. Click `fabricbeat-*`, then the star in the upper right corner:
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Index_pattern_selection_basic.png "Setting default index pattern")


### About indices

Three different Elasticsearch indices per Fabric organization are setup. One for blocks, one for transactions and one for single writes.  If multiple agents are run for peers in the same organization, they are going to send their data to the same indices. You can then select the peer on the dashboards to view its data only.  
If multiple instances are run for peers in different organizations, you will see the data of different organizations on different dashboards.  

The name of the indices can be customized in the fabricbeat configuration file (\_meta/beat.yml and `make update` or directly in fabricbeat.yml).


## Viewing dashboards

After that, we can click the dashboards and see the overview of our data on the Overview Dashboard (org1):
![alt text](https://github.com/hyperledger-labs/blockchain-analyzer/blob/master/docs/images/Overview_with_filter_multi.png "Overview with filter")
**If the dashboards are empty, set the time range wider!**

We can go on discovering the dashboards by scrolling and clicking the link fields, or by selecting another dashboards from the Dashboards menu.


## Starting more instances of fabricbeat agent
To start more instances of the fabricbeat agent, open another tab/terminal, make sure that the GOPATH variable is set (`export GOPATH=$HOME/go`) , and run fabricbeat passing different variables from the previous run(s) (e.g.
```
ORG_NUMBER=2 PEER_NUMBER=0 NETWORK=multichannel ./fabricbeat -e -d "*"
```
will start an agent querying peer0.org2.el-network.com). If the started instance queries a peer from the same organization as the previous one, we can select the peer we want to see the data of from a dropdown on the dashboards. If the new peer is shipping data from a different organization, we can see its data on a different dashboard (click the dashboards menu on the left, and choose one).
